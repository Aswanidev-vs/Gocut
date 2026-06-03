package app

import (
	"context"
	"encoding/base64"
	"fmt"
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

// App is the root application object bound to Wails frontend.
type App struct {
	ctx            context.Context
	executor       *ffmpeg.Executor
	importer       *media.Importer
	projectMgr     *project.Manager
	store          *project.Store
	renderQueue    *render.Queue
	compositor     *render.Compositor
	thumbCache     *cache.ThumbnailCache
	frameCache     *cache.FrameCache
	fontScanner    *fonts.Scanner
	autosave       *project.AutoSaver
	currentProject *project.Project
	ffmpegPath     string
	ffprobePath    string
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
	if err := a.executor.ExtractThumbnail(a.ctx, asset.Path, float64(timeMs)/1000.0, outPath); err != nil {
		return "", err
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		return "", err
	}

	_ = a.thumbCache.Put(assetID+".png", data)

	a.emit("asset:thumbnailReady", project.AssetThumbnailEvent{
		AssetID: assetID,
		Data:    base64.StdEncoding.EncodeToString(data),
	})

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

	a.emit("asset:waveformReady", project.AssetWaveformEvent{
		AssetID: assetID,
		Data:    wf,
	})

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

func (a *App) GetPreviewFrame(p project.Project, timeSeconds float64, width int, height int) (string, error) {
	asset, sourceTime := a.assetAtTime(p, timeSeconds)
	if asset == nil {
		return "", fmt.Errorf("no asset at time %.2fs", timeSeconds)
	}

	cacheKey := fmt.Sprintf("preview/%s/%.3f.jpg", asset.ID, sourceTime)

	if cached, ok := a.thumbCache.Get(cacheKey); ok {
		return base64.StdEncoding.EncodeToString(cached), nil
	}

	data, err := a.executor.ExtractThumbnailPipe(a.ctx, asset.Path, sourceTime)
	if err != nil {
		return "", fmt.Errorf("thumbnail extraction failed: %w", err)
	}

	_ = a.thumbCache.Put(cacheKey, data)

	return base64.StdEncoding.EncodeToString(data), nil
}

func (a *App) PreloadFrames(p project.Project, startTime float64, count int) error {
	for i := 0; i < count; i++ {
		t := startTime + float64(i)*(1.0/30.0)
		asset, sourceTime := a.assetAtTime(p, t)
		if asset == nil {
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
	for i := range p.Timeline.Tracks {
		track := &p.Timeline.Tracks[i]
		if track.Type != project.TrackVideo {
			continue
		}
		for j := range track.Clips {
			c := &track.Clips[j]
			if t >= c.StartTime && t < c.StartTime+c.Duration {
				for k := range p.Assets {
					if p.Assets[k].ID == c.AssetID {
						src := c.TrimStart + (t - c.StartTime)
						return &p.Assets[k], src
					}
				}
			}
		}
	}
	for i := range p.Assets {
		if p.Assets[i].Type == project.AssetVideo {
			return &p.Assets[i], t
		}
	}
	return nil, 0
}

// --- Utility ---

func (a *App) CheckFFmpegInstalled() (string, error) {
	if a.ffmpegPath == "" {
		a.emit("ffmpeg:notFound", nil)
		return "", fmt.Errorf("ffmpeg not found")
	}
	if a.ffprobePath == "" {
		return "", fmt.Errorf("ffprobe not found")
	}
	cmd := exec.Command(a.ffmpegPath, "-version")
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

func (a *App) GetAppVersion() string {
	return "0.1.0"
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
	case fileExists("xdg-open"):
		cmd = exec.Command("xdg-open", path)
	case fileExists("explorer.exe"):
		cmd = exec.Command("explorer.exe", path)
	default:
		return fmt.Errorf("no file opener found")
	}
	return cmd.Start()
}

func fileExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func joinPatterns(exts []string) string {
	out := ""
	for i, e := range exts {
		if i > 0 {
			out += ";"
		}
		out += "*." + e
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

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}
	a.mediaServerPort = listener.Addr().(*net.TCPAddr).Port
	go http.Serve(listener, mux)
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
