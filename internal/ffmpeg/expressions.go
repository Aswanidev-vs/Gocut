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
			// linear interpolation
			lerp = fmt.Sprintf("(%g+(%g-%g)*(t-%g)/%g)", v1, v2, v1, t1, t2-t1)
		}

		nextExpr := buildExpr(idx + 1)
		return fmt.Sprintf("if(lt(t,%g),%s,%s)", t2, lerp, nextExpr)
	}

	firstKF := kfs[0]
	return fmt.Sprintf("if(lt(t,%g),%g,%s)", firstKF.Time, getFloat(firstKF.Value), buildExpr(0))
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
