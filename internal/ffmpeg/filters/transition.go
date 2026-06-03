package filters

import (
	"fmt"

	"Gocut/internal/project"
)

func BuildTransitionFilter(t project.Transition, offset float64, idxA, idxB int) string {
	switch t.Type {
	case "fade":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=fade:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "dissolve":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=dissolve:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "wipeleft":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=wipeleft:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "wiperight":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=wiperight:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "slideleft":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=slideleft:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "slideright":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=slideright:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "zoomin":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=zoomin:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "hflip":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=hflip:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "circleopen":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=circleopen:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "pixelize":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=pixelize:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	case "blur":
		return fmt.Sprintf("[v%d][v%d]xfade=transition=blur:duration=%g:offset=%g[vout]", idxA, idxB, t.Duration, offset)
	default:
		return ""
	}
}
