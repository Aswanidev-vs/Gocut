package ffmpeg

import (
	"fmt"
	"math"
	"sort"

	"Gocut/internal/project"
)

// BuildAnimatedExpression builds an FFmpeg mathematical expression for a specific property.
// It interpolates between keyframes based on local clip time 't'.
func BuildAnimatedExpression(keyframes []project.Keyframe, property string, baseValue float64) string {
	return buildAnimatedExpression(keyframes, property, baseValue, "t")
}

// BuildAnimatedExpressionT is like BuildAnimatedExpression, but emits the
// uppercase time variable "T". It is meant for filters such as geq that
// evaluate their own per-frame time variable — lowercase t is undefined
// there and makes FFmpeg reject the filtergraph at parse time.
func BuildAnimatedExpressionT(keyframes []project.Keyframe, property string, baseValue float64) string {
	return buildAnimatedExpression(keyframes, property, baseValue, "T")
}

func buildAnimatedExpression(keyframes []project.Keyframe, property string, baseValue float64, timeVar string) string {
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
			ratio := buildEasingRatioExpr(t1, t2, kf1.Easing, timeVar)
			lerp = fmt.Sprintf("(%g+(%g-%g)*%s)", v1, v2, v1, ratio)
		}

		nextExpr := buildExpr(idx + 1)
		return fmt.Sprintf("if(lt(%s,%g),%s,%s)", timeVar, t2, lerp, nextExpr)
	}

	firstKF := kfs[0]
	return fmt.Sprintf("if(lt(%s,%g),%g,%s)", timeVar, firstKF.Time, getFloat(firstKF.Value), buildExpr(0))
}

// buildEasingRatioExpr returns an FFmpeg expression fragment that maps local
// time (timeVar in [t1,t2]) to an interpolation ratio. bounce and elastic
// mirror applyEasing() in frontend/src/stores/designStore.js, so the canvas
// preview and the exported video agree. All numeric constants are folded on
// the Go side so FFmpeg only ever sees plain decimal literals.
func buildEasingRatioExpr(t1, t2 float64, easing string, timeVar string) string {
	dur := t2 - t1
	u := fmt.Sprintf("((%s-%g)/%g)", timeVar, t1, dur)
	switch easing {
	case "easeIn", "ease-in":
		return fmt.Sprintf("(%s*%s)", u, u)
	case "easeOut", "ease-out":
		return fmt.Sprintf("(%s*(2-%s))", u, u)
	case "easeInOut", "ease-in-out":
		return fmt.Sprintf("if(lt(%s,0.5),2*%s*%s,1-pow(-2*%s+2,2)/2)", u, u, u, u)
	case "bounce", "easeOutBounce":
		// bounceOut exactly as the frontend computes it:
		//   n1 = 7.5625, d1 = 2.75
		//   u < 1/d1    -> n1*u*u
		//   u < 2/d1    -> n1*(u-1.5/d1)^2  + 0.75
		//   u < 2.5/d1  -> n1*(u-2.25/d1)^2 + 0.9375
		//   else        -> n1*(u-2.625/d1)^2 + 0.984375
		const n1 = 7.5625
		const d1 = 2.75
		return fmt.Sprintf(
			"if(lt(%[1]s,%[2]s),%[3]s*%[1]s*%[1]s,"+
				"if(lt(%[1]s,%[4]s),%[3]s*(%[1]s-%[5]s)*(%[1]s-%[5]s)+0.75,"+
				"if(lt(%[1]s,%[6]s),%[3]s*(%[1]s-%[7]s)*(%[1]s-%[7]s)+0.9375,"+
				"%[3]s*(%[1]s-%[8]s)*(%[1]s-%[8]s)+0.984375)))",
			u,
			fmt.Sprintf("%g", 1/d1),     // 0.363636...
			fmt.Sprintf("%g", n1),       // 7.5625
			fmt.Sprintf("%g", 2/d1),     // 0.727272...
			fmt.Sprintf("%g", 1.5/d1),   // 0.545454...
			fmt.Sprintf("%g", 2.5/d1),   // 0.909090...
			fmt.Sprintf("%g", 2.25/d1),  // 0.818181...
			fmt.Sprintf("%g", 2.625/d1), // 0.954545...
		)
	case "elastic", "elasticOut", "easeOutElastic":
		// -2^(10u-10) * sin((u*10-10.75) * 2*PI/3), with exact endpoints.
		// 2*PI/3 is precomputed in Go (≈ 2.0943951023931953). The pi
		// variable forces float64 (double) arithmetic so the constant
		// matches the JS frontend double rounding bit-for-bit; the
		// untyped constant expression 2*math.Pi/3 would be folded at
		// extended precision by the Go compiler and differ in the last ULP.
		pi := math.Pi
		c := fmt.Sprintf("%g", 2*pi/3)
		// lte(u,0) clamps the exact endpoint, mirroring the JS
		// `t === 0 || t === 1` branch.
		return fmt.Sprintf("if(lte(%[1]s,0),0,if(lt(%[1]s,1),"+
			"-pow(2,10*%[1]s-10)*sin((%[1]s*10-10.75)*%[2]s),1))", u, c)
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
