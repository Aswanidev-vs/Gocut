package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// IsImageExt returns true if the given filename has an image extension
// supported by the image demuxer in ffmpeg.
func IsImageExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp":
		return true
	}
	return false
}

// ExtractImageThumbnail decodes a still image and writes the first frame
// to outputPath. We use the image2 demuxer with -loop 1 so a single
// synthetic frame is available to ffmpeg regardless of the source format.
func (e *Executor) ExtractImageThumbnail(ctx context.Context, inputPath string, outputPath string) error {
	args := []string{
		"-loop", "1",
		"-i", inputPath,
		"-frames:v", "1",
		"-q:v", "2",
		"-y",
		outputPath,
	}
	out, err := e.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("image thumbnail extraction failed: %w: %s", err, string(out))
	}
	return nil
}

// ExtractImageThumbnailPipe streams a still image's only frame to stdout
// as a single mjpeg image. This is the fastest path to a preview frame for
// image assets because it avoids any seek into a single-frame demuxer.
func (e *Executor) ExtractImageThumbnailPipe(ctx context.Context, inputPath string) ([]byte, error) {
	args := []string{
		"-loop", "1",
		"-i", inputPath,
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-",
	}
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	PrepareCmd(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ExtractThumbnail writes a single JPEG frame of the input to outputPath.
// For image inputs the timestamp is forced to 0 and the image2 demuxer is
// used with -loop 1 so that seeking beyond the single synthetic frame does
// not fail.
func (e *Executor) ExtractThumbnail(ctx context.Context, inputPath string, timeSec float64, outputPath string) error {
	var args []string
	if IsImageExt(inputPath) {
		// Image demuxer: loop the single frame indefinitely so seek/time arguments
		// are well-defined, and ignore any non-zero timestamp by clamping to 0.
		args = []string{
			"-loop", "1",
			"-i", inputPath,
			"-frames:v", "1",
			"-q:v", "2",
			"-y",
			outputPath,
		}
	} else {
		args = []string{
			"-ss", floatToStr(timeSec),
			"-i", inputPath,
			"-frames:v", "1",
			"-q:v", "2",
			"-y",
			outputPath,
		}
	}
	out, err := e.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("thumbnail extraction failed: %w: %s", err, string(out))
	}
	return nil
}

// ExtractThumbnailPipe streams a single JPEG frame to stdout. The same image
// demuxer treatment as ExtractThumbnail is applied.
func (e *Executor) ExtractThumbnailPipe(ctx context.Context, inputPath string, timeSec float64) ([]byte, error) {
	var args []string
	if IsImageExt(inputPath) {
		args = []string{
			"-loop", "1",
			"-i", inputPath,
			"-frames:v", "1",
			"-f", "image2pipe",
			"-vcodec", "mjpeg",
			"-",
		}
	} else {
		args = []string{
			"-ss", floatToStr(timeSec),
			"-i", inputPath,
			"-frames:v", "1",
			"-f", "image2pipe",
			"-vcodec", "mjpeg",
			"-",
		}
	}
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	PrepareCmd(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func GenerateThumbnailStrip(ctx context.Context, exe *Executor, inputPath string, count int) ([]string, error) {
	if count <= 0 {
		return nil, nil
	}

	info, err := exe.Probe(ctx, inputPath)
	if err != nil {
		return nil, err
	}

	duration := info.Duration
	if duration <= 0 {
		return nil, nil
	}

	step := duration / float64(count+1)
	thumbs := make([]string, 0, count)

	for i := 1; i <= count; i++ {
		t := step * float64(i)
		ext := filepath.Ext(inputPath)
		base := inputPath[:len(inputPath)-len(ext)]
		outPath := base + "_thumb_" + strconv.Itoa(i) + ".jpg"

		if err := exe.ExtractThumbnail(ctx, inputPath, t, outPath); err != nil {
			return thumbs, err
		}
		thumbs = append(thumbs, outPath)
	}

	return thumbs, nil
}

func floatToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', 3, 64)
}

func intToStr(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	out := ""
	for i > 0 {
		out = string(rune('0'+i%10)) + out
		i /= 10
	}
	if neg {
		out = "-" + out
	}
	return out
}
