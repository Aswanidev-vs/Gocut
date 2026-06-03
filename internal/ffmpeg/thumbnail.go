package ffmpeg

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
)

func (e *Executor) ExtractThumbnail(ctx context.Context, inputPath string, timeSec float64, outputPath string) error {
	args := []string{
		"-ss", floatToStr(timeSec),
		"-i", inputPath,
		"-frames:v", "1",
		"-q:v", "2",
		"-y",
		outputPath,
	}
	out, err := e.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("thumbnail extraction failed: %w: %s", err, string(out))
	}
	return nil
}

func (e *Executor) ExtractThumbnailPipe(ctx context.Context, inputPath string, timeSec float64) ([]byte, error) {
	args := []string{
		"-ss", floatToStr(timeSec),
		"-i", inputPath,
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "mjpeg",
		"-",
	}
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
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
