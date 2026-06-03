package filters

import (
	"fmt"
	"strings"
)

func BuildAudioFilterChain(volume float64, fadeIn, fadeOut bool, fadeDuration, endTime float64) string {
	var parts []string
	if volume != 1.0 {
		parts = append(parts, fmt.Sprintf("volume=%g", volume))
	}
	if fadeIn {
		parts = append(parts, fmt.Sprintf("afade=t=in:st=0:d=%g", fadeDuration))
	}
	if fadeOut {
		st := endTime - fadeDuration
		if st < 0 {
			st = 0
		}
		parts = append(parts, fmt.Sprintf("afade=t=out:st=%g:d=%g", st, fadeDuration))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ",")
}
