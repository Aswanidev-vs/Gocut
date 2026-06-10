//go:build windows

package ffmpeg

import (
	"os/exec"
	"syscall"
)

// PrepareCmd hides the console window for the given command on Windows.
func PrepareCmd(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}
