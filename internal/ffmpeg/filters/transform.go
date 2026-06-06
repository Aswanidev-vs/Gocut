package filters

import (
	"fmt"
	"strings"

	"Gocut/internal/ffmpeg"
	"Gocut/internal/project"
)

func BuildTransformFilter(clip project.Clip) string {
	var parts []string
	t := clip.Transform

	rotExpr := ffmpeg.BuildAnimatedExpression(clip.Keyframes, "rotation", t.Rotation)
	if rotExpr != "0" {
		parts = append(parts, fmt.Sprintf("rotate='%s*PI/180'", rotExpr))
	}
	
	scaleXExpr := ffmpeg.BuildAnimatedExpression(clip.Keyframes, "scaleX", t.ScaleX)
	scaleYExpr := ffmpeg.BuildAnimatedExpression(clip.Keyframes, "scaleY", t.ScaleY)
	
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
