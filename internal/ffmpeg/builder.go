package ffmpeg

import (
	"fmt"
	"strconv"
	"strings"

	"Gocut/internal/project"
)

type FilterGraph struct {
	parts []string
}

func NewFilterGraph() *FilterGraph {
	return &FilterGraph{}
}

func (g *FilterGraph) Add(expr string) *FilterGraph {
	g.parts = append(g.parts, expr)
	return g
}

func (g *FilterGraph) Build() string {
	return strings.Join(g.parts, ";")
}

func BuildTransformFilters(clip project.Clip) string {
	var parts []string
	t := clip.Transform

	rotExpr := BuildAnimatedExpression(clip.Keyframes, "rotation", t.Rotation)
	if rotExpr != "0" {
		parts = append(parts, fmt.Sprintf("rotate='%s*PI/180:c=black@0:eval=frame'", rotExpr))
	}

	scaleXExpr := BuildAnimatedExpression(clip.Keyframes, "scaleX", t.ScaleX)
	scaleYExpr := BuildAnimatedExpression(clip.Keyframes, "scaleY", t.ScaleY)
	if scaleXExpr != "1" || scaleYExpr != "1" || t.ScaleX != 1 || t.ScaleY != 1 {
		parts = append(parts, fmt.Sprintf("scale=w='iw*%s':h='ih*%s':eval=frame", scaleXExpr, scaleYExpr))
	}

	if t.CropW > 0 && t.CropH > 0 {
		parts = append(parts, fmt.Sprintf("crop=%g:%g:%g:%g", t.CropW, t.CropH, t.CropX, t.CropY))
	}
	if t.FlipH {
		parts = append(parts, "hflip")
	}
	if t.FlipV {
		parts = append(parts, "vflip")
	}
	return strings.Join(parts, ",")
}

// BuildAtempoChain returns atempo filter(s) for a playback speed factor.
// FFmpeg's atempo only accepts 0.5–2.0 per instance, so rates outside that
// range are expressed as a chain of segments (e.g. 4x -> atempo=2.0,atempo=2.0).
// Returns "" when speed is 0 (no change) or ~1.0 (identity).
func BuildAtempoChain(speed float64) string {
	if speed <= 0 || (speed >= 0.9999 && speed <= 1.0001) {
		return ""
	}
	remaining := speed
	var parts []string
	for remaining > 1.0001 {
		step := remaining
		if step > 2.0 {
			step = 2.0
		}
		parts = append(parts, fmt.Sprintf("atempo=%g", step))
		remaining /= step
	}
	for remaining < 0.9999 {
		step := remaining
		if step < 0.5 {
			step = 0.5
		}
		parts = append(parts, fmt.Sprintf("atempo=%g", step))
		remaining /= step
	}
	return strings.Join(parts, ",")
}

func BuildAudioFilters(volume float64, fadeIn, fadeOut bool, fadeDuration, endTime float64) string {
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
	return strings.Join(parts, ",")
}

func BuildLoudNormFilter() string {
	return "loudnorm=I=-16:TP=-1.5:LRA=11"
}

func BuildNoiseReductionFilter() string {
	return "afftdn=nf=-25"
}

func BuildTextFilter(p project.TextProps) string {
	if p.Text == "" {
		return ""
	}
	filter := "drawtext=text='" + escapeText(p.Text) + "'"
	if p.FontFamily != "" {
		filter += ":fontfile='" + p.FontFamily + "'"
	}
	filter += fmt.Sprintf(":fontsize=%d", p.FontSize)
	if p.Color != "" {
		filter += ":fontcolor=" + p.Color
	}
	return filter
}

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

func escapeText(s string) string {
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, ":", "\\:")
	return s
}

// FloatToStr renders a float for use inside FFmpeg numeric expressions
// (-ss, scale factors, keyframe values). The previous rational
// approximation was wrong for fractions (0.5 -> "0/2" == 0),
// so we emit a plain decimal float, which FFmpeg parses correctly.
func FloatToStr(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}
