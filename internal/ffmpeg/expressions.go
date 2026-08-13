package ffmpeg

import (
	"fmt"
	"sort"

	"Gocut/internal/project"
)

// BuildAnimatedExpression builds an FFmpeg mathematical expression for a specific property.
// It uses linear interpolation between keyframes based on local clip time 't'.
func BuildAnimatedExpression(keyframes []project.Keyframe, property string, baseValue float64) string {
	var kfs []project.Keyframe
	for _, kf := range keyframes {
		if kf.Property == property {
			kfs = append(kfs, kf)
		}
	}

	if len(kfs) == 0 {
		return FloatToStr(baseValue)
	}

	// Sort keyframes by time
	sort.Slice(kfs, func(i, j int) bool {
		return kfs[i].Time < kfs[j].Time
	})

	// Recursive function to build the nested if/lerp expression
	var buildExpr func(idx int) string
	buildExpr = func(idx int) string {
		if idx >= len(kfs)-1 {
			return FloatToStr(getFloat(kfs[idx].Value))
		}
		
		kf1 := kfs[idx]
		kf2 := kfs[idx+1]
		t1 := kf1.Time
		t2 := kf2.Time
		v1 := getFloat(kf1.Value)
		v2 := getFloat(kf2.Value)

		var lerp string
		if t2 == t1 {
			lerp = FloatToStr(v2)
		} else {
			ratio := buildEasingRatioExpr(t1, t2, kf1.Easing)
			lerp = fmt.Sprintf("(%g+(%g-%g)*%s)", v1, v2, v1, ratio)
		}

		nextExpr := buildExpr(idx + 1)
		return fmt.Sprintf("if(lt(t,%g),%s,%s)", t2, lerp, nextExpr)
	}

	firstKF := kfs[0]
	return fmt.Sprintf("if(lt(t,%g),%g,%s)", firstKF.Time, getFloat(firstKF.Value), buildExpr(0))
}

func buildEasingRatioExpr(t1, t2 float64, easing string) string {
	dur := t2 - t1
	u := fmt.Sprintf("((t-%g)/%g)", t1, dur)
	switch easing {
	case "easeIn", "ease-in":
		return fmt.Sprintf("(%s*%s)", u, u)
	case "easeOut", "ease-out":
		return fmt.Sprintf("(%s*(2-%s))", u, u)
	case "easeInOut", "ease-in-out":
		return fmt.Sprintf("if(lt(%s,0.5),2*%s*%s,1-pow(-2*%s+2,2)/2)", u, u, u, u)
	default:
		return u
	}
}

func getFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case float32:
		return float64(val)
	default:
		return 0
	}
}
