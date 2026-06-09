//go:build !windows

package ffmpeg

import (
	"os/exec"
)

// hideWindowOnWindows is a no-op on non-Windows platforms. macOS and
// Linux do not allocate a hidden console for child processes, so the
// CREATE_NO_WINDOW flag is unnecessary there.
func hideWindowOnWindows(cmd *exec.Cmd) {
	_ = cmd
}
