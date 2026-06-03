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
	return cmd.CombinedOutput()
}

func (e *Executor) RunBackground(ctx context.Context, args ...string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
