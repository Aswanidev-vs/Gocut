package filters

import (
	"fmt"
	"strings"

	"Gocut/internal/project"
)

func BuildColorFilterChain(c project.ColorGrade) string {
	var parts []string

	eqReady := false
	var eq []string

	addEq := func(s string) {
		if !eqReady {
			eqReady = true
		}
		eq = append(eq, s)
	}

	if c.Brightness != 0 {
		addEq(fmt.Sprintf("eq=brightness=%g", float64(c.Brightness)/100.0))
	}
	if c.Contrast != 0 {
		addEq(fmt.Sprintf("contrast=%g", 1.0+float64(c.Contrast)/100.0))
	}
	if c.Saturation != 0 {
		addEq(fmt.Sprintf("saturation=%g", 1.0+float64(c.Saturation)/100.0))
	}

	if len(eq) > 0 {
		parts = append(parts, strings.Join(eq, ":"))
	}

	if c.Hue != 0 {
		parts = append(parts, fmt.Sprintf("hue=h=%d", c.Hue))
	}

	if needsColorBalance(c) {
		parts = append(parts, buildColorBalance(c))
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
	if strings.TrimSpace(c.Curves) != "" {
		parts = append(parts, fmt.Sprintf("curves=%s", c.Curves))
	}
	if c.ChromaKeyColor != "" {
		similarity := c.ChromaKeySimilarity
		if similarity <= 0 {
			similarity = 0.01 // ffmpeg default
		}
		blend := c.ChromaKeyBlend
		// Convert CSS hex (#00ff00) to FFmpeg format (0x00ff00)
		color := c.ChromaKeyColor
		if strings.HasPrefix(color, "#") {
			color = "0x" + strings.TrimPrefix(color, "#")
		}
		parts = append(parts, fmt.Sprintf("chromakey=color=%s:similarity=%g:blend=%g", color, similarity, blend))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, ",")
}

func needsColorBalance(c project.ColorGrade) bool {
	return c.Temp != 0 || c.Tint != 0 || c.Highlights != 0 || c.Shadows != 0 ||
		c.LiftR != 0 || c.LiftG != 0 || c.LiftB != 0 ||
		c.GammaR != 0 || c.GammaG != 0 || c.GammaB != 0 ||
		c.GainR != 0 || c.GainG != 0 || c.GainB != 0
}

func buildColorBalance(c project.ColorGrade) string {
	var parts []string

	var rh, rh_neg, gh, gh_neg, bh, bh_neg string

	// Lift maps to shadow channel offsets
	rh_neg = offsetStr(c.LiftR)
	gh_neg = offsetStr(c.LiftG)
	bh_neg = offsetStr(c.LiftB)

	// Gamma maps to midtones
	rh = offsetStr(c.GammaR)
	gh = offsetStr(c.GammaG)
	bh = offsetStr(c.GammaB)

	// Gain maps to highlights
	// highlights/shadows act as separate scalars shadowed onto rh_neg/bh_neg and rh/bh
	// keep it simple and pack into the same channel offsets
	if c.Highlights != 0 {
		rh = addOffset(rh, c.Highlights)
		gh = addOffset(gh, c.Highlights)
		bh = addOffset(bh, c.Highlights)
	}
	if c.Shadows != 0 {
		rh_neg = addOffset(rh_neg, c.Shadows)
		gh_neg = addOffset(gh_neg, c.Shadows)
		bh_neg = addOffset(bh_neg, c.Shadows)
	}

	if c.Temp != 0 {
		// positive temp = warm, negative = cool
		rh = addOffset(rh, c.Temp)
		bh = addOffset(bh, -c.Temp)
	}
	if c.Tint != 0 {
		gh = addOffset(gh, c.Tint)
		// simultaneously nudge red/blue opposite to keep luminance roughly stable
		rh = addOffset(rh, -c.Tint/2)
		bh = addOffset(bh, -c.Tint/2)
	}

	// Base gain mapping only if anything nonzero
	if c.GainR != 0 || c.GainG != 0 || c.GainB != 0 {
		if rh == "" { rh = "0" }
		if gh == "" { gh = "0" }
		if bh == "" { bh = "0" }
	}

	if rh != "" { parts = append(parts, "rh="+rh) }
	if gh != "" { parts = append(parts, "gh="+gh) }
	if bh != "" { parts = append(parts, "bh="+bh) }
	if rh_neg != "" { parts = append(parts, "rh_neg="+rh_neg) }
	if gh_neg != "" { parts = append(parts, "gh_neg="+gh_neg) }
	if bh_neg != "" { parts = append(parts, "bh_neg="+bh_neg) }

	if len(parts) == 0 {
		return ""
	}
	return "colorbalance=" + strings.Join(parts, ":")
}

func offsetStr(v int) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%g", float64(v)/100.0)
}

func addOffset(base string, v int) string {
	if v == 0 {
		return base
	}
	s := offsetStr(v)
	if base == "" {
		return s
	}
	return fmt.Sprintf("%s+%g", base, float64(v)/100.0)
}
