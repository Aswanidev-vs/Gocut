package ffmpeg

import (
	"context"
	"os/exec"
)

type Executor struct {
	ffmpegPath string
	ffprobePath string
}

func NewExecutor(ffmpegPath string, ffprobePath string) *Executor {
	return &Executor{ffmpegPath: ffmpegPath, ffprobePath: ffprobePath}
}

func (e *Executor) FFmpegPath() string {
	return e.ffmpegPath
}

func (e *Executor) FFprobePath() string {
	return e.ffprobePath
}

func (e *Executor) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	PrepareCmd(cmd)
	return cmd.CombinedOutput()
}

// RunOut runs ffmpeg and returns only stdout. Use it when the output is a
// binary payload (e.g. a piped preview frame) — CombinedOutput would mix
// stderr logs into the stream and corrupt it.
func (e *Executor) RunOut(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	PrepareCmd(cmd)
	return cmd.Output()
}

func (e *Executor) RunBackground(ctx context.Context, args ...string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	PrepareCmd(cmd)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
