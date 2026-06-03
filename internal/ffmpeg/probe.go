package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"Gocut/internal/project"
)

func (e *Executor) Probe(ctx context.Context, path string) (*project.MediaInfo, error) {
	if e.ffprobePath == "" {
		return nil, fmt.Errorf("ffprobe not found")
	}

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		path,
	}
	cmd := exec.CommandContext(ctx, e.ffprobePath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w: %s", err, string(out))
	}

	var probe struct {
		Format struct {
			Duration string `json:"duration"`
			Size     string `json:"size"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
		Streams []struct {
			CodecType  string `json:"codec_type"`
			CodecName  string `json:"codec_name"`
			Width      int    `json:"width"`
			Height     int    `json:"height"`
			RFrameRate string `json:"r_frame_rate"`
		} `json:"streams"`
	}

	if err := json.Unmarshal(out, &probe); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	info := &project.MediaInfo{Path: path}

	if s, err := strconv.ParseInt(probe.Format.Size, 10, 64); err == nil {
		info.FileSize = s
	}

	dur, _ := strconv.ParseFloat(probe.Format.Duration, 64)
	info.Duration = dur

	for _, s := range probe.Streams {
		if s.CodecType == "video" {
			info.Width = s.Width
			info.Height = s.Height
			info.Codec = s.CodecName
			info.FPS = parseFPS(s.RFrameRate)
		}
		if s.CodecType == "audio" {
			info.AudioCodec = s.CodecName
		}
	}

	return info, nil
}

func parseFPS(expr string) float64 {
	parts := strings.Split(expr, "/")
	if len(parts) != 2 {
		return 0
	}
	num, _ := strconv.ParseFloat(parts[0], 64)
	den, _ := strconv.ParseFloat(parts[1], 64)
	if den == 0 {
		return 0
	}
	return num / den
}
