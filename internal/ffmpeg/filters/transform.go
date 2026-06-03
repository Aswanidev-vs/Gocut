package filters

import (
	"fmt"
	"strings"

	"Gocut/internal/project"
)

func BuildTransformFilter(t project.Transform) string {
	var parts []string
	if t.Rotation != 0 {
		parts = append(parts, fmt.Sprintf("rotate=%g*PI/180", t.Rotation))
	}
	parts = append(parts, fmt.Sprintf("scale=%g:%g", t.ScaleX, t.ScaleY))
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
