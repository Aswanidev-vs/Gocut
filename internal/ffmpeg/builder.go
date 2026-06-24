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
		parts = append(parts, fmt.Sprintf("rotate='%s*PI/180:c=black@0'", rotExpr))
	}

	scaleXExpr := BuildAnimatedExpression(clip.Keyframes, "scaleX", t.ScaleX)
	scaleYExpr := BuildAnimatedExpression(clip.Keyframes, "scaleY", t.ScaleY)
	if scaleXExpr != "1" || scaleYExpr != "1" || t.ScaleX != 1 || t.ScaleY != 1 {
		parts = append(parts, fmt.Sprintf("scale=w='iw*%s':h='ih*%s':eval=frame", scaleXExpr, scaleYExpr))
	}

	if t.CropW > 0 && t.CropH > 0 {
		parts = append(parts, fmt.Sprintf("crop=iw*%g:ih*%g:iw*%g:ih*%g", t.CropW, t.CropH, t.CropX, t.CropY))
	}
	if t.FlipH {
		parts = append(parts, "hflip")
	}
	if t.FlipV {
		parts = append(parts, "vflip")
	}
	return strings.Join(parts, ",")
}

func BuildColorFilters(c project.ColorGrade) string {
	var parts []string
	if c.Brightness != 0 || c.Contrast != 0 || c.Saturation != 0 {
		parts = append(parts, fmt.Sprintf("eq=brightness=%g:contrast=%g:saturation=%g",
			float64(c.Brightness)/100.0,
			float64(c.Contrast)/100.0,
			float64(c.Saturation)/100.0))
	}
	if c.Hue != 0 {
		parts = append(parts, fmt.Sprintf("hue=h=%d", c.Hue))
	}
	if c.Sharpness > 0 {
		parts = append(parts, fmt.Sprintf("unsharp=5:5:%d", c.Sharpness))
	}
	if c.Vignette > 0 {
		parts = append(parts, fmt.Sprintf("vignette=angle=PI/4*%d", c.Vignette))
	}
	if c.Grain > 0 {
		parts = append(parts, fmt.Sprintf("noise=alls=%d", c.Grain))
	}
	if c.Blur > 0 {
		parts = append(parts, fmt.Sprintf("boxblur=%d", c.Blur))
	}
	if c.ChromaKeyColor != "" {
		similarity := c.ChromaKeySimilarity
		if similarity <= 0 {
			similarity = 0.01
		}
		blend := c.ChromaKeyBlend
		color := c.ChromaKeyColor
		if strings.HasPrefix(color, "#") {
			color = "0x" + strings.TrimPrefix(color, "#")
		}
		parts = append(parts, fmt.Sprintf("chromakey=color=%s:similarity=%g:blend=%g", color, similarity, blend))
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

func FloatToStr(f float64) string {
	if f == 0 {
		return "0"
	}
	s := ""
	if f < 0 {
		s += "-"
		f = -f
	}
	s += strconv.Itoa(int(f))
	frac := f - float64(int(f))
	if frac > 0 {
		s += "/" + strconv.Itoa(int(1/frac))
	}
	return s
}

func IntToStr(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	out := ""
	for i > 0 {
		out = string(rune('0'+i%10)) + out
		i /= 10
	}
	if neg {
		out = "-" + out
	}
	return out
}
