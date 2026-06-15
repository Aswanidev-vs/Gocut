package app

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"Gocut/internal/cache"
	"Gocut/internal/ffmpeg"
	"Gocut/internal/fonts"
	"Gocut/internal/media"
	"Gocut/internal/project"
	"Gocut/internal/render"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Version is set at build time via -ldflags "-X 'Gocut/internal/app.Version=x.y.z'"
var Version = "dev"

// audioExtracting tracks in-progress audio extractions to avoid duplicate ffmpeg runs.
var audioExtracting sync.Map

// App is the root application object bound to Wails frontend.
type App struct {
	ctx             context.Context
	executor        *ffmpeg.Executor
	importer        *media.Importer
	projectMgr      *project.Manager
	store           *project.Store
	renderQueue     *render.Queue
	compositor      *render.Compositor
	thumbCache      *cache.ThumbnailCache
	frameCache      *cache.FrameCache
	fontScanner     *fonts.Scanner
	autosave        *project.AutoSaver
	currentProject  *project.Project
	ffmpegPath      string
	ffprobePath     string
	prefetchMu      sync.Mutex
	prefetchAsset   string
	prefetchTime    float64
	mediaServerPort int
}

func NewApp() *App {
	return &App{}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.ffmpegPath = findFFmpeg()
	a.ffprobePath = findFFprobe()
	a.executor = ffmpeg.NewExecutor(a.ffmpegPath, a.ffprobePath)
	a.importer = media.NewImporter(a.executor)

	storePath := appDataDir("gocut.db")
	store, err := project.NewStore(storePath)
	if err != nil {
		store, _ = project.NewStore(":memory:")
	}
	a.store = store
	a.projectMgr = project.NewManager(a.store)

	a.renderQueue = render.NewQueue(ctx)
	a.compositor = render.NewCompositor(a.executor)

	thumbDir := appDataDir("thumbnails")
	thumbCache, _ := cache.NewThumbnailCache(thumbDir, 500*1024*1024)
	a.thumbCache = thumbCache

	a.frameCache = cache.NewFrameCache(120, 30*time.Second)
	a.fontScanner = fonts.NewScanner()
	a.autosave = project.NewAutoSaver(a.projectMgr, 60)
	a.autosave.SetProjectSupplier(func() *project.Project { return a.currentProject })
	a.autosave.Start(ctx)
	a.startMediaServer()
}

func (a *App) emit(event string, data interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event, data)
	}
}

// --- Project ---

func (a *App) NewProject(settings project.ProjectSettings) (*project.Project, error) {
	p, err := a.projectMgr.NewProject(settings)
	if err != nil {
		return nil, err
	}
	a.currentProject = p
	return p, nil
}

func (a *App) SaveProject(p project.Project) error {
	if a.currentProject != nil && a.currentProject.FilePath != "" && a.currentProject.FilePath != p.FilePath {
		if fileExists(a.currentProject.FilePath) {
			_ = os.Rename(a.currentProject.FilePath, p.FilePath)
		}
	}
	a.currentProject = &p
	return a.projectMgr.SaveProject(p)
}

func (a *App) LoadProject(path string) (*project.Project, error) {
	p, err := a.projectMgr.LoadProject(path)
	if err != nil {
		return nil, err
	}
	a.currentProject = p
	return p, nil
}

func (a *App) GetRecentProjects() ([]project.RecentProject, error) {
	if a.projectMgr == nil {
		return []project.RecentProject{}, nil
	}
	return a.projectMgr.GetRecentProjects(10)
}

func (a *App) DeleteProject(id string) error {
	if a.projectMgr == nil {
		return nil
	}
	return a.projectMgr.DeleteProject(id)
}

func (a *App) ClearRecentProjects() error {
	if a.projectMgr == nil {
		return nil
	}
	return a.projectMgr.ClearRecent()
}

func (a *App) ExportProjectFile(p project.Project) (string, error) {
	return a.projectMgr.ExportProjectFile(p)
}

// --- Media ---

func (a *App) ImportMedia(paths []string) ([]project.Asset, error) {
	assets, err := a.importer.ImportMedia(a.ctx, paths)
	if err != nil {
		return nil, err
	}
	if a.currentProject != nil && len(assets) > 0 {
		a.currentProject.Assets = append(a.currentProject.Assets, assets...)
	}
	return assets, nil
}

func (a *App) ExtractThumbnail(assetID string, timeMs int) (string, error) {
	asset := a.findAsset(assetID)
	if asset == nil {
		return "", fmt.Errorf("asset not found: %s", assetID)
	}

	thumbDir, err := ensureAppDataSubdir("thumbnails")
	if err != nil {
		return "", err
	}
	outPath := filepath.Join(thumbDir, assetID+".png")

	// For still images there is no seek; just decode the first frame.
	if asset.Type == project.AssetImage {
		if err := a.executor.ExtractImageThumbnail(a.ctx, asset.Path, outPath); err != nil {
			return "", err
		}
	} else {
		if err := a.executor.ExtractThumbnail(a.ctx, asset.Path, float64(timeMs)/1000.0, outPath); err != nil {
			return "", err
		}
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		return "", err
	}

	_ = a.thumbCache.Put(assetID+".png", data)

	go func() {
		a.emit("asset:thumbnailReady", project.AssetThumbnailEvent{
			AssetID: assetID,
			Data:    base64.StdEncoding.EncodeToString(data),
		})
	}()

	return base64.StdEncoding.EncodeToString(data), nil
}

func (a *App) ExtractWaveform(assetID string) ([]float32, error) {
	asset := a.findAsset(assetID)
	if asset == nil {
		return nil, fmt.Errorf("asset not found: %s", assetID)
	}

	wf, err := a.executor.ExtractWaveform(a.ctx, asset.Path, 500)
	if err != nil {
		return nil, err
	}

	go func() {
		a.emit("asset:waveformReady", project.AssetWaveformEvent{
			AssetID: assetID,
			Data:    wf,
		})
	}()

	return wf, nil
}

func (a *App) GetMediaInfo(path string) (*project.MediaInfo, error) {
	return a.executor.Probe(a.ctx, path)
}

func (a *App) GenerateThumbnailStrip(assetID string, count int) ([]string, error) {
	asset := a.findAsset(assetID)
	if asset == nil {
		return nil, fmt.Errorf("asset not found: %s", assetID)
	}
	return ffmpeg.GenerateThumbnailStrip(a.ctx, a.executor, asset.Path, count)
}

// --- Render ---

func (a *App) StartRender(p project.Project, settings project.RenderSettings) (string, error) {
	job := &render.Job{
		ID:         "",
		Project:    p,
		Settings:   settings,
		OutputPath: settings.OutputPath,
		Status:     "queued",
	}
	jobID := a.renderQueue.Enqueue(job)
	return jobID, nil
}

func (a *App) GetRenderProgress(jobID string) (*project.RenderProgress, error) {
	job := a.renderQueue.Get(jobID)
	if job == nil {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}
	return &project.RenderProgress{
		JobID:   job.ID,
		Percent: job.Progress,
		Status:  job.Status,
		Error:   errToString(job.Error),
	}, nil
}

func (a *App) CancelRender(jobID string) error {
	return a.renderQueue.Cancel(jobID)
}

func (a *App) GetRenderQueue() ([]project.RenderProgress, error) {
	jobs := a.renderQueue.List()
	var out []project.RenderProgress
	for _, j := range jobs {
		out = append(out, project.RenderProgress{
			JobID:   j.ID,
			Percent: j.Progress,
			Status:  j.Status,
		})
	}
	return out, nil
}

func (a *App) OpenOutputFolder(path string) error {
	if path == "" {
		return nil
	}
	folder := filepath.Dir(path)
	return openFolder(folder)
}

// --- Preview ---

// resolveVisualAsset returns the asset that should be rendered at the
// given timeline time, and a "source time" within that asset.
// For stills the source time is forced to 0 so that ffmpeg's image2
// demuxer always succeeds.
func (a *App) resolveVisualAsset(p project.Project, t float64) (*project.Asset, float64) {
	if asset, src := a.assetAtTime(p, t); asset != nil {
		return asset, src
	}
	// Fallback: if a still image exists in the project but no clip
	// happens to cover t, return the first image so the preview is at
	// least a meaningful placeholder rather than a black frame.
	for i := range p.Assets {
		if p.Assets[i].Type == project.AssetImage {
			return &p.Assets[i], 0
		}
	}
	return nil, 0
}

func (a *App) GetPreviewFrame(p project.Project, timeSeconds float64, width int, height int) (string, error) {
	asset, sourceTime := a.resolveVisualAsset(p, timeSeconds)
	if asset == nil {
		return "", fmt.Errorf("no asset at time %.2fs", timeSeconds)
	}

	// Image assets are single-frame stills; cache by asset only.
	cacheKey := fmt.Sprintf("preview/%s/%.3f.jpg", asset.ID, sourceTime)
	if asset.Type == project.AssetImage {
		cacheKey = fmt.Sprintf("preview/%s/img.jpg", asset.ID)
	}

	if cached, ok := a.thumbCache.Get(cacheKey); ok {
		return base64.StdEncoding.EncodeToString(cached), nil
	}

	var data []byte
	var err error
	if asset.Type == project.AssetImage {
		data, err = a.executor.ExtractImageThumbnailPipe(a.ctx, asset.Path)
	} else {
		data, err = a.executor.ExtractThumbnailPipe(a.ctx, asset.Path, sourceTime)
	}
	if err != nil {
		return "", fmt.Errorf("thumbnail extraction failed: %w", err)
	}

	_ = a.thumbCache.Put(cacheKey, data)

	return base64.StdEncoding.EncodeToString(data), nil
}

func (a *App) PreloadFrames(p project.Project, startTime float64, count int) error {
	for i := 0; i < count; i++ {
		t := startTime + float64(i)*(1.0/30.0)
		asset, sourceTime := a.resolveVisualAsset(p, t)
		if asset == nil {
			continue
		}
		// Skip preloading for stills: there is only one frame and
		// GetPreviewFrame will have already cached it.
		if asset.Type == project.AssetImage {
			continue
		}
		cacheKey := fmt.Sprintf("preview/%s/%.3f.jpg", asset.ID, sourceTime)
		if _, ok := a.thumbCache.Get(cacheKey); ok {
			continue
		}
		go func(inputPath string, st float64, key string) {
			_ = a.executor.ExtractThumbnail(a.ctx, inputPath, st, filepath.Join(appDataDir("preview"), key))
		}(asset.Path, sourceTime, cacheKey)
	}
	return nil
}

func (a *App) ClearPreviewCache() error {
	a.frameCache.Clear()
	_ = a.thumbCache.Clear(nil)
	return nil
}

func (a *App) assetAtTime(p project.Project, t float64) (*project.Asset, float64) {
	// Priority list of visual tracks to check first
	visualTypes := []project.TrackType{project.TrackVideo, project.TrackImage, project.TrackPIP}
	for _, trackType := range visualTypes {
		for i := range p.Timeline.Tracks {
			track := &p.Timeline.Tracks[i]
			if track.Type != trackType {
				continue
			}
			for j := range track.Clips {
				c := &track.Clips[j]
				if t >= c.StartTime && t < c.StartTime+c.Duration {
					for k := range p.Assets {
						if p.Assets[k].ID == c.AssetID {
							asset := &p.Assets[k]
							// Still images have no timeline; always seek to 0 so
							// ffmpeg's image2 demuxer can produce a frame.
							if asset.Type == project.AssetImage {
								return asset, 0
							}
							src := c.TrimStart + (t - c.StartTime)
							return asset, src
						}
					}
				}
			}
		}
	}

	// Fallback to first video asset ONLY if timeline has no visual clips at all
	hasAnyVisualClip := false
	for i := range p.Timeline.Tracks {
		track := &p.Timeline.Tracks[i]
		if track.Type == project.TrackVideo || track.Type == project.TrackImage || track.Type == project.TrackPIP {
			if len(track.Clips) > 0 {
				hasAnyVisualClip = true
				break
			}
		}
	}
	if !hasAnyVisualClip {
		for i := range p.Assets {
			if p.Assets[i].Type == project.AssetVideo {
				if p.Assets[i].Duration > 0 && t >= p.Assets[i].Duration {
					t = p.Assets[i].Duration - 0.001
				}
				return &p.Assets[i], t
			}
		}
	}
	return nil, 0
}

// --- Utility ---

func (a *App) CheckFFmpegInstalled() (string, error) {
	if a.ffmpegPath == "" {
		go func() {
			a.emit("ffmpeg:notFound", nil)
		}()
		return "", fmt.Errorf("ffmpeg not found")
	}
	if a.ffprobePath == "" {
		return "", fmt.Errorf("ffprobe not found")
	}
	cmd := exec.Command(a.ffmpegPath, "-version")
	ffmpeg.PrepareCmd(cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}
	return "ffmpeg", nil
}

// OpenDirectoryPicker uses Wails OpenDirectoryDialog to select a folder.
func (a *App) OpenDirectoryPicker() (string, error) {
	options := runtime.OpenDialogOptions{
		Title: "Select Save Folder",
	}
	path, err := runtime.OpenDirectoryDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	return path, nil
}

// OpenFilePicker uses the Wails multi-file dialog so the JS side can rely
// on receiving a []string (instead of having to special-case a single path
// or an empty string). The previous implementation returned at most one
// element, which made the frontend's `Array.isArray` check always fail.
func (a *App) OpenFilePicker(filters []project.FileFilter) ([]string, error) {
	options := runtime.OpenDialogOptions{
		Title:   "Select files",
		Filters: []runtime.FileFilter{},
	}
	for _, f := range filters {
		patterns := joinPatterns(f.Extensions)
		options.Filters = append(options.Filters, runtime.FileFilter{
			DisplayName: f.Name,
			Pattern:     patterns,
		})
	}
	paths, err := runtime.OpenMultipleFilesDialog(a.ctx, options)
	if err != nil {
		return nil, err
	}
	return paths, nil
}

// SaveFilePicker uses the Wails save file dialog for choosing an export output location.
func (a *App) SaveFilePicker(defaultName string, filters []project.FileFilter) (string, error) {
	options := runtime.SaveDialogOptions{
		Title:           "Save Export File",
		DefaultFilename: defaultName,
		Filters:         []runtime.FileFilter{},
	}
	for _, f := range filters {
		patterns := joinPatterns(f.Extensions)
		options.Filters = append(options.Filters, runtime.FileFilter{
			DisplayName: f.Name,
			Pattern:     patterns,
		})
	}
	path, err := runtime.SaveFileDialog(a.ctx, options)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) GetAppVersion() string {
	return Version
}

func (a *App) Minimise() {
	if a.ctx != nil {
		runtime.WindowMinimise(a.ctx)
	}
}

func (a *App) Maximise() {
	if a.ctx != nil {
		runtime.WindowMaximise(a.ctx)
	}
}

func (a *App) Close() {
	if a.ctx != nil {
		runtime.Quit(a.ctx)
	}
}

// --- Private helpers ---

func (a *App) findAsset(id string) *project.Asset {
	if a.currentProject == nil {
		return nil
	}
	for i := range a.currentProject.Assets {
		if a.currentProject.Assets[i].ID == id {
			return &a.currentProject.Assets[i]
		}
	}
	return nil
}

func errToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func openFolder(path string) error {
	var cmd *exec.Cmd
	switch {
	case fileExistsOnPath("xdg-open"):
		cmd = exec.Command("xdg-open", path)
	case fileExistsOnPath("explorer.exe"):
		cmd = exec.Command("explorer.exe", path)
	default:
		return fmt.Errorf("no file opener found")
	}
	return cmd.Start()
}

// fileExistsOnPath checks if an *executable* exists in $PATH (for ffmpeg, etc.)
func fileExistsOnPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// fileExists checks if a *file path* exists on disk.
func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

func joinPatterns(exts []string) string {
	out := ""
	for i, e := range exts {
		if i > 0 {
			out += ";"
		}
		lower := strings.ToLower(e)
		upper := strings.ToUpper(e)
		out += "*." + lower + ";*." + upper
	}
	return out
}

func appDataDir(name string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	dir := filepath.Join(home, ".gocut")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, name)
}

func ensureAppDataSubdir(name string) (string, error) {
	path := appDataDir(name)
	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	return path, nil
}

// --- Media file server ---

func (a *App) startMediaServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/media", func(w http.ResponseWriter, r *http.Request) {
		assetPath := r.URL.Query().Get("path")
		if assetPath == "" {
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}
		// Security: only serve files that exist as assets in the current project.
		if !a.isKnownAssetPath(assetPath) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Range")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Range, Content-Length, Accept-Ranges")
		f, err := os.Open(assetPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.ServeContent(w, r, filepath.Base(assetPath), stat.ModTime(), f)
	})

	mux.HandleFunc("/audio", func(w http.ResponseWriter, r *http.Request) {
		assetPath := r.URL.Query().Get("path")
		if assetPath == "" {
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}
		if !a.isKnownAssetPath(assetPath) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		cacheDir, _ := ensureAppDataSubdir("audio_cache")
		pathBytes := []byte(assetPath)
		// SHA-256 (rather than MD5) is used here purely as a stable
		// cache-key derivation; it is not a security primitive, but
		// using a modern hash also keeps static-analysis tooling happy.
		hash := fmt.Sprintf("%x", sha256.Sum256(pathBytes))
		cachedPath := filepath.Join(cacheDir, hash+".mp3")

		if _, statErr := os.Stat(cachedPath); os.IsNotExist(statErr) {
			doneCh := make(chan struct{})
			actual, loaded := audioExtracting.LoadOrStore(cachedPath, doneCh)
			if loaded {
				select {
				case <-actual.(chan struct{}):
				case <-r.Context().Done():
					http.Error(w, "cancelled", http.StatusRequestTimeout)
					return
				}
			} else {
				defer func() {
					audioExtracting.Delete(cachedPath)
					close(doneCh)
				}()
				if a.ffmpegPath == "" {
					http.Error(w, "ffmpeg not found", http.StatusInternalServerError)
					return
				}
				cmd := exec.CommandContext(r.Context(), a.ffmpegPath,
					"-y", "-i", assetPath,
					"-vn",
					"-c:a", "libmp3lame", "-q:a", "4",
					"-ar", "44100",
					cachedPath,
				)
				ffmpeg.PrepareCmd(cmd)
				if err := cmd.Run(); err != nil {
					_ = os.Remove(cachedPath)
					http.Error(w, "audio extraction failed: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Range")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Range, Content-Length, Accept-Ranges")
		f, err := os.Open(cachedPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		stat, _ := f.Stat()
		http.ServeContent(w, r, filepath.Base(cachedPath), stat.ModTime(), f)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}
	a.mediaServerPort = listener.Addr().(*net.TCPAddr).Port

	server := &http.Server{
		Handler:  mux,
		ErrorLog: log.New(io.Discard, "", 0),
	}
	go server.Serve(listener)
}

func (a *App) isKnownAssetPath(path string) bool {
	if a.currentProject == nil {
		return false
	}
	for _, asset := range a.currentProject.Assets {
		if asset.Path == path {
			return true
		}
	}
	return false
}

func (a *App) GetMediaServerPort() int {
	return a.mediaServerPort
}

func findFFmpeg() string {
	for _, name := range []string{"ffmpeg", "ffmpeg.exe"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}

func findFFprobe() string {
	for _, name := range []string{"ffprobe", "ffprobe.exe"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	return ""
}
