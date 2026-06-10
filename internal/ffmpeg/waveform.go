package ffmpeg

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os/exec"
	"strconv"
)

func (e *Executor) ExtractWaveform(ctx context.Context, inputPath string, samples int) ([]float32, error) {
	if samples <= 0 {
		samples = 500
	}

	args := []string{
		"-i", inputPath,
		"-filter_complex", "aformat=channel_layouts=mono,compand,aresample=" + strconv.Itoa(samples),
		"-ac", "1",
		"-f", "f32le",
		"-",
	}

	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	PrepareCmd(cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("waveform extraction failed: %w", err)
	}

	if len(out) == 0 {
		return []float32{}, nil
	}

	count := len(out) / 4
	if count > samples {
		count = samples
	}

	data := make([]float32, count)
	for i := 0; i < count; i++ {
		bits := binary.LittleEndian.Uint32(out[i*4 : (i+1)*4])
		data[i] = math.Float32frombits(bits)
	}

	return normalizeWaveform(data), nil
}

func normalizeWaveform(data []float32) []float32 {
	if len(data) == 0 {
		return data
	}

	max := float32(0)
	for _, v := range data {
		abs := v
		if abs < 0 {
			abs = -abs
		}
		if abs > max {
			max = abs
		}
	}

	if max == 0 {
		return data
	}

	out := make([]float32, len(data))
	for i, v := range data {
		out[i] = v / max
	}
	return out
}
