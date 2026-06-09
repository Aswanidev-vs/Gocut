package ffmpeg

import (
	"context"
	"os/exec"
)

type Executor struct {
	ffmpegPath  string
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

// runFFmpeg is the single chokepoint for spawning ffmpeg.exe. Every
// ffmpeg call in the codebase should go through it so that the
// platform-specific CREATE_NO_WINDOW flag is applied uniformly.
func (e *Executor) runFFmpeg(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	hideWindowOnWindows(cmd)
	return cmd
}

// runFFprobe is the single chokepoint for spawning ffprobe.exe.
func (e *Executor) runFFprobe(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, e.ffprobePath, args...)
	hideWindowOnWindows(cmd)
	return cmd
}

// RunWithHiddenWindowProgram is a package-level helper used by the
// render queue, which builds its ffmpeg argv manually with the
// program name in args[0]. It applies the same CREATE_NO_WINDOW
// flag on Windows so the ffmpeg child process never pops a console
// window.
func RunWithHiddenWindowProgram(ctx context.Context, program string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, program, args...)
	hideWindowOnWindows(cmd)
	return cmd
}

func (e *Executor) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := e.runFFmpeg(ctx, args...)
	return cmd.CombinedOutput()
}

func (e *Executor) RunBackground(ctx context.Context, args ...string) (*exec.Cmd, error) {
	cmd := e.runFFmpeg(ctx, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
