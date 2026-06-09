package media

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Gocut/internal/ffmpeg"
	"Gocut/internal/project"
)

type Importer struct {
	executor *ffmpeg.Executor
}

func NewImporter(exe *ffmpeg.Executor) *Importer {
	return &Importer{executor: exe}
}

var supportedVideoExts = map[string]bool{
	".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true, ".ts": true, ".m4v": true,
}
var supportedAudioExts = map[string]bool{
	".mp3": true, ".wav": true, ".aac": true, ".flac": true, ".ogg": true, ".m4a": true, ".wma": true,
}
var supportedImageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true, ".bmp": true,
}

// ImportResult reports what was successfully imported and which files failed.
// Returning a non-nil error here means *every* file failed; partial failures
// are reported via the Skipped slice while Assets contains what did work.
type ImportResult struct {
	Assets  []project.Asset
	Skipped []SkippedFile
}

type SkippedFile struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// ImportMedia imports every supported file. If a single file fails the rest
// still continue, so the user can drag a folder of mixed media and see
// everything that did work plus a toast listing the rejects.
func (i *Importer) ImportMedia(ctx context.Context, paths []string) ([]project.Asset, error) {
	var assets []project.Asset
	var skipped []SkippedFile

	for _, p := range paths {
		asset, err := i.importSingle(ctx, p)
		if err != nil {
			skipped = append(skipped, SkippedFile{Path: p, Reason: err.Error()})
			continue
		}
		assets = append(assets, *asset)
	}

	// Only return a hard error if literally nothing succeeded.
	if len(assets) == 0 && len(skipped) > 0 {
		return assets, fmt.Errorf("no media could be imported (%d files skipped)", len(skipped))
	}
	return assets, nil
}

// probeImage pulls width/height directly from an image file via Go's stdlib
// image decoder. ffprobe does not expose a video stream for stills
// (PNG/JPG/WebP/BMP), so a regular probe returns Width=Height=0 which
// breaks downstream rendering and aspect-ratio calculations.
func probeImage(ctx context.Context, path string) (width, height int, err error) {
	type result struct {
		w, h int
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		f, openErr := os.Open(path)
		if openErr != nil {
			ch <- result{err: openErr}
			return
		}
		defer f.Close()
		cfg, _, decErr := image.DecodeConfig(f)
		if decErr != nil {
			ch <- result{err: decErr}
			return
		}
		ch <- result{w: cfg.Width, h: cfg.Height}
	}()
	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	case r := <-ch:
		return r.w, r.h, r.err
	}
}

func (i *Importer) importSingle(ctx context.Context, path string) (*project.Asset, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if !isSupported(ext) {
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	assetType := detectAssetType(ext, nil)

	info, err := i.executor.Probe(ctx, path)
	if err != nil {
		// Still images have no streams; ffprobe may exit non-zero. Treat
		// this as a soft failure for image types and continue with the
		// Go-side image decoder.
		if assetType != project.AssetImage {
			return nil, fmt.Errorf("probe failed: %w", err)
		}
		info = &project.MediaInfo{Path: path}
	}

	// For images, ffprobe returns no video stream and Width/Height come
	// back as 0. Decode the image header with the stdlib to recover the
	// real dimensions so the compositor can place the still correctly.
	width := info.Width
	height := info.Height
	if assetType == project.AssetImage && (width == 0 || height == 0) {
		if w, h, derr := probeImage(ctx, path); derr == nil {
			width = w
			height = h
		}
	}

	// Images have no intrinsic duration; default to 5s so the clip
	// shows up on the timeline at a usable length. The user can trim.
	duration := info.Duration
	if assetType == project.AssetImage && duration <= 0 {
		duration = 5.0
	}

	asset := &project.Asset{
		ID:         generateID(),
		Path:       path,
		Type:       assetType,
		Duration:   duration,
		Width:      width,
		Height:     height,
		FPS:        info.FPS,
		Codec:      info.Codec,
		FileSize:   info.FileSize,
		ImportedAt: time.Now(),
	}

	return asset, nil
}

func (i *Importer) PopulateWaveform(ctx context.Context, asset *project.Asset) error {
	if asset.Type != project.AssetVideo && asset.Type != project.AssetAudio {
		return nil
	}
	wf, err := i.executor.ExtractWaveform(ctx, asset.Path, 500)
	if err != nil {
		return err
	}
	asset.Waveform = wf
	return nil
}

func isSupported(ext string) bool {
	return supportedVideoExts[ext] || supportedAudioExts[ext] || supportedImageExts[ext]
}

func detectAssetType(ext string, info *project.MediaInfo) project.AssetType {
	switch ext {
	case ".mp3", ".wav", ".aac", ".flac", ".ogg", ".m4a", ".wma":
		return project.AssetAudio
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
		return project.AssetImage
	}
	return project.AssetVideo
}

func generateID() string {
	return fmt.Sprintf("asset_%d", time.Now().UnixNano())
}

func (i *Importer) ValidatePath(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if !isSupported(ext) {
		return fmt.Errorf("unsupported format: %s", ext)
	}
	if _, err := os.Stat(path); err != nil {
		return err
	}
	return nil
}
