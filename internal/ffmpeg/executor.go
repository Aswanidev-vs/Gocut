package ffmpeg

import (
	"context"
	"os/exec"
	"syscall"
)

// hideWindowOnWindows ensures the spawned child process does not pop up a
// console window on Windows. On macOS / Linux this is a no-op.
//
// The default Windows behaviour when a Go process (even a windowsgui one)
// launches a child with os/exec is to attach / create a console for the
// child if the child itself is a console-subsystem binary — which ffmpeg,
// ffprobe and explorer.exe are. That results in a "cmd.exe" window
// flashing behind the GUI for every thumbnail / waveform / render call.
//
// We side-step this by setting CREATE_NO_WINDOW (0x08000000) in the
// process creation flags. The child still runs, just without spawning
// a visible window.
func hideWindowOnWindows(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// CREATE_NO_WINDOW = 0x08000000
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= 0x08000000
}

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
// CREATE_NO_WINDOW flag is applied uniformly.
func (e *Executor) runFFmpeg(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	hideWindowOnWindows(cmd)
	return cmd
}

// runFFprobe is the single chokepoint for spawning ffprobe.exe. It
// applies CREATE_NO_WINDOW so the helper console does not flash
// behind the editor on Windows.
func (e *Executor) runFFprobe(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, e.ffprobePath, args...)
	hideWindowOnWindows(cmd)
	return cmd
}

// RunWithHiddenWindow spawns an arbitrary executable with arguments
// while still applying the CREATE_NO_WINDOW flag. This is used by the
// render queue, which builds its ffmpeg argv manually rather than
// going through runFFmpeg (because the program path comes from
// buildSimpleFFmpegArgs, not from the executor's ffmpegPath field).
func (e *Executor) RunWithHiddenWindow(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, e.ffmpegPath, args...)
	hideWindowOnWindows(cmd)
	return cmd
}

// RunWithHiddenWindowProgram is like RunWithHiddenWindow but takes an
// explicit program path (used by the render queue where the program
// path is encoded into args[0]).
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
