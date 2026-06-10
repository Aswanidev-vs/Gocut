//go:build !windows

package ffmpeg

import "os/exec"

// PrepareCmd is a no-op on non-Windows platforms.
func PrepareCmd(cmd *exec.Cmd) {
	// No-op
}
