package media

import (
	"context"
	"os/exec"

	"Gocut/internal/ffmpeg"
	"Gocut/internal/project"
)

type MediaInfoFetcher struct {
	executor *ffmpeg.Executor
}

func NewMediaInfoFetcher(exe *ffmpeg.Executor) *MediaInfoFetcher {
	return &MediaInfoFetcher{executor: exe}
}

func (m *MediaInfoFetcher) GetMediaInfo(ctx context.Context, path string) (*project.MediaInfo, error) {
	info, err := m.executor.Probe(ctx, path)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (m *MediaInfoFetcher) ExtractThumbnailBase64(ctx context.Context, path string, timeMs int) (string, error) {
	args := []string{
		"-ss", ffmpeg.FloatToStr(float64(timeMs)/1000.0),
		"-i", path,
		"-frames:v", "1",
		"-f", "image2pipe",
		"-vcodec", "png",
		"-",
	}

	cmd := exec.CommandContext(ctx, m.executor.FFmpegPath(), args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return "data:image/png;base64," + encodeBase64(out), nil
}

func encodeBase64(b []byte) string {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var dst []byte
	for i := 0; i < len(b); i += 3 {
		b1 := b[i]
		var b2, b3 byte
		if i+1 < len(b) {
			b2 = b[i+1]
		}
		if i+2 < len(b) {
			b3 = b[i+2]
		}
		dst = append(dst, alphabet[b1>>2])
		dst = append(dst, alphabet[((b1&0x03)<<4)|(b2>>4)])
		if i+1 >= len(b) {
			dst = append(dst, '=', '=', '=')
			break
		}
		dst = append(dst, alphabet[((b2&0x0f)<<2)|(b3>>6)])
		if i+2 >= len(b) {
			dst = append(dst, '=')
			break
		}
		dst = append(dst, alphabet[b3&0x3f])
	}
	return string(dst)
}
