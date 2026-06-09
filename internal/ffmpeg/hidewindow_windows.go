//go:build windows

package ffmpeg

import (
	"os/exec"
	"syscall"
)

// hideWindowOnWindows sets the CREATE_NO_WINDOW flag on the spawned
// child process so that a console-subsystem binary (like ffmpeg.exe
// or ffprobe.exe) does not flash a cmd.exe window behind the Gocut
// editor every time the user imports, previews, or renders media.
//
// This file is Windows-only. The non-Windows implementation in
// hidewindow_other.go is a no-op.
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
