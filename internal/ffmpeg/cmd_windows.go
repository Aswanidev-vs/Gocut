//go:build windows

package ffmpeg

import (
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// PrepareCmd hides the console window for the given command on Windows.
func PrepareCmd(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.Env = sanitizedFFmpegEnv(cmd.Env)
}

// Some Windows FFmpeg builds link against Fontconfig but do not ship a usable
// default config. If the parent process inherits broken FONTCONFIG_* variables,
// FFmpeg can fail even when drawtext uses an explicit font file.
func sanitizedFFmpegEnv(existing []string) []string {
	env := existing
	if len(env) == 0 {
		env = os.Environ()
	}

	out := make([]string, 0, len(env))
	for _, entry := range env {
		key := entry
		if idx := strings.IndexByte(entry, '='); idx >= 0 {
			key = entry[:idx]
		}
		switch strings.ToUpper(key) {
		case "FONTCONFIG_FILE", "FONTCONFIG_PATH":
			continue
		default:
			out = append(out, entry)
		}
	}
	return out
}
