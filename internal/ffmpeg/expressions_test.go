package ffmpeg

import (
	"math"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"Gocut/internal/project"
)

// ---- minimal evaluator for the subset of FFmpeg av_expr that
// BuildAnimatedExpression emits, so the generated strings can be checked
// numerically without launching ffmpeg.

type exprParser struct {
	s    string
	pos  int
	vars map[string]float64
}

func parseExprAndEval(s string, vars map[string]float64) (float64, error) {
	p := &exprParser{s: s, vars: vars}
	v, err := p.parseExpr()
	if err != nil {
		return 0, err
	}
	p.skipSpaces()
	if p.pos != len(p.s) {
		return 0, &errParse{msg: "trailing input at " + strconv.Itoa(p.pos) + " in " + s}
	}
	return v, nil
}

type errParse struct{ msg string }

func (e *errParse) Error() string { return e.msg }

func (p *exprParser) skipSpaces() {
	for p.pos < len(p.s) && (p.s[p.pos] == ' ' || p.s[p.pos] == '\t') {
		p.pos++
	}
}

func (p *exprParser) peek() byte {
	p.skipSpaces()
	if p.pos < len(p.s) {
		return p.s[p.pos]
	}
	return 0
}

func (p *exprParser) accept(c byte) bool {
	p.skipSpaces()
	if p.pos < len(p.s) && p.s[p.pos] == c {
		p.pos++
		return true
	}
	return false
}

func (p *exprParser) parseExpr() (float64, error) {
	v, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		c := p.peek()
		if c == '+' {
			p.pos++
			r, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v += r
		} else if c == '-' {
			p.pos++
			r, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			v -= r
		} else {
			return v, nil
		}
	}
}

func (p *exprParser) parseTerm() (float64, error) {
	v, err := p.parseUnary()
	if err != nil {
		return 0, err
	}
	for {
		c := p.peek()
		if c == '*' {
			p.pos++
			r, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			v *= r
		} else if c == '/' {
			p.pos++
			r, err := p.parseUnary()
			if err != nil {
				return 0, err
			}
			v /= r
		} else {
			return v, nil
		}
	}
}

func (p *exprParser) parseUnary() (float64, error) {
	c := p.peek()
	if c == '-' {
		p.pos++
		v, err := p.parseUnary()
		return -v, err
	}
	if c == '+' {
		p.pos++
		return p.parseUnary()
	}
	return p.parsePrimary()
}

func (p *exprParser) parsePrimary() (float64, error) {
	p.skipSpaces()
	if p.pos >= len(p.s) {
		return 0, &errParse{msg: "unexpected end of input"}
	}
	c := p.s[p.pos]
	if c == '(' {
		p.pos++
		v, err := p.parseExpr()
		if err != nil {
			return 0, err
		}
		if !p.accept(')') {
			return 0, &errParse{msg: "missing )"}
		}
		return v, nil
	}
	if c == '.' || (c >= '0' && c <= '9') {
		start := p.pos
		for p.pos < len(p.s) {
			ch := p.s[p.pos]
			if (ch >= '0' && ch <= '9') || ch == '.' || ch == 'e' || ch == 'E' ||
				((ch == '+' || ch == '-') && p.pos > start && (p.s[p.pos-1] == 'e' || p.s[p.pos-1] == 'E')) {
				p.pos++
			} else {
				break
			}
		}
		return strconv.ParseFloat(p.s[start:p.pos], 64)
	}
	if unicode.IsLetter(rune(c)) || c == '_' {
		start := p.pos
		for p.pos < len(p.s) && (unicode.IsLetter(rune(p.s[p.pos])) || p.s[p.pos] == '_') {
			p.pos++
		}
		name := p.s[start:p.pos]
		if p.accept('(') { // function call
			var args []float64
			if p.peek() != ')' {
				for {
					a, err := p.parseExpr()
					if err != nil {
						return 0, err
					}
					args = append(args, a)
					if p.accept(',') {
						continue
					}
					break
				}
			}
			if !p.accept(')') {
				return 0, &errParse{msg: "missing ) after " + name}
			}
			return p.callFunc(name, args)
		}
		if v, ok := p.vars[name]; ok {
			return v, nil
		}
		switch name {
		case "PI":
			return math.Pi, nil
		case "E":
			return math.E, nil
		}
		return 0, &errParse{msg: "undefined identifier " + name}
	}
	return 0, &errParse{msg: "unexpected character " + string(c)}
}

func (p *exprParser) callFunc(name string, args []float64) (float64, error) {
	switch name {
	case "if":
		if len(args) != 3 {
			return 0, &errParse{msg: "if wants 3 args"}
		}
		if args[0] > 0 {
			return args[1], nil
		}
		return args[2], nil
	case "lt":
		if len(args) != 2 {
			return 0, &errParse{msg: "lt wants 2 args"}
		}
		if args[0] < args[1] {
			return 1, nil
		}
		return 0, nil
	case "lte":
		if len(args) != 2 {
			return 0, &errParse{msg: "lte wants 2 args"}
		}
		if args[0] <= args[1] {
			return 1, nil
		}
		return 0, nil
	case "pow":
		return math.Pow(args[0], args[1]), nil
	case "sin":
		return math.Sin(args[0]), nil
	case "cos":
		return math.Cos(args[0]), nil
	case "abs":
		return math.Abs(args[0]), nil
	case "min":
		return math.Min(args[0], args[1]), nil
	case "max":
		return math.Max(args[0], args[1]), nil
	case "clip":
		if args[0] < args[1] {
			return args[1], nil
		}
		if args[0] > args[2] {
			return args[2], nil
		}
		return args[0], nil
	default:
		return 0, &errParse{msg: "unknown function " + name}
	}
}

func evalAt(t *testing.T, expr string, time float64) float64 {
	t.Helper()
	v, err := parseExprAndEval(expr, map[string]float64{"t": time, "T": time})
	if err != nil {
		t.Fatalf("eval %q at t=%v: %v", expr, time, err)
	}
	return v
}

// Reference easings — must stay in sync with applyEasing() in
// frontend/src/stores/designStore.js.

func refEaseIn(u float64) float64 { return u * u }

func refEaseOut(u float64) float64 { return u * (2 - u) }

func refEaseInOut(u float64) float64 {
	if u < 0.5 {
		return 2 * u * u
	}
	return 1 - math.Pow(-2*u+2, 2)/2
}

func refBounce(u float64) float64 {
	const n1, d1 = 7.5625, 2.75
	switch {
	case u < 1/d1:
		return n1 * u * u
	case u < 2/d1:
		d := u - 1.5/d1
		return n1*d*d + 0.75
	case u < 2.5/d1:
		d := u - 2.25/d1
		return n1*d*d + 0.9375
	default:
		d := u - 2.625/d1
		return n1*d*d + 0.984375
	}
}

func refElastic(u float64) float64 {
	if u <= 0 {
		return 0
	}
	if u >= 1 {
		return 1
	}
	return -math.Pow(2, 10*u-10) * math.Sin((u*10-10.75)*(2*math.Pi/3))
}

// ---- structure tests ----

func TestBuildAnimatedExpressionNoKeyframes(t *testing.T) {
	got := BuildAnimatedExpression(nil, "x", 42.5)
	if got != "42.5" {
		t.Fatalf("got %q, want %q", got, "42.5")
	}
	if v := evalAt(t, got, 3); v != 42.5 {
		t.Fatalf("eval = %v, want 42.5", v)
	}
}

func TestBuildAnimatedExpressionIgnoresOtherProperties(t *testing.T) {
	kfs := []project.Keyframe{
		{ID: "k1", Time: 0, Property: "y", Value: 99.0},
	}
	got := BuildAnimatedExpression(kfs, "x", 10)
	if got != "10" {
		t.Fatalf("got %q, want constant 10 (no x keyframes)", got)
	}
}

func TestBuildAnimatedExpressionSingleKeyframe(t *testing.T) {
	kfs := []project.Keyframe{
		{ID: "k1", Time: 1.5, Property: "x", Value: 20.0},
	}
	got := BuildAnimatedExpression(kfs, "x", 0)
	// Before and after the only keyframe the value is constant.
	for _, time := range []float64{0, 1.0, 1.5, 5} {
		if v := evalAt(t, got, time); v != 20 {
			t.Fatalf("at t=%v got %v, want 20", time, v)
		}
	}
}

func TestBuildAnimatedExpressionLinearMidpoint(t *testing.T) {
	kfs := []project.Keyframe{
		{ID: "k1", Time: 0, Property: "x", Value: 0.0},
		{ID: "k2", Time: 2, Property: "x", Value: 10.0},
	}
	got := BuildAnimatedExpression(kfs, "x", 0)
	cases := []struct{ time, want float64 }{
		{0, 0},
		{0.5, 2.5},
		{1, 5},
		{1.5, 7.5},
		{2, 10},
		{3, 10}, // after last keyframe: constant
	}
	for _, c := range cases {
		if v := evalAt(t, got, c.time); math.Abs(v-c.want) > 1e-9 {
			t.Errorf("t=%v: got %v, want %v", c.time, v, c.want)
		}
	}
}

func TestBuildAnimatedExpressionEaseIn(t *testing.T) {
	kfs := []project.Keyframe{
		{ID: "k1", Time: 0, Property: "x", Value: 0.0, Easing: "easeIn"},
		{ID: "k2", Time: 2, Property: "x", Value: 10.0},
	}
	got := BuildAnimatedExpression(kfs, "x", 0)
	// easeIn at u=0.5 (t=1) yields ratio 0.25 -> value 2.5
	if v := evalAt(t, got, 1); math.Abs(v-2.5) > 1e-9 {
		t.Fatalf("at t=1 got %v, want 2.5 (easeIn u^2 at u=0.5)", v)
	}
}

func TestBuildAnimatedExpressionClampBeforeFirstKeyframe(t *testing.T) {
	kfs := []project.Keyframe{
		{ID: "k1", Time: 1, Property: "x", Value: 5.0},
		{ID: "k2", Time: 3, Property: "x", Value: 15.0},
	}
	got := BuildAnimatedExpression(kfs, "x", 999)
	if v := evalAt(t, got, 0); v != 5 {
		t.Fatalf("before first keyframe: got %v, want 5", v)
	}
}

// ---- easing parity tests: generated expression vs reference easing ----

func ratioTestKfs(easing string) []project.Keyframe {
	return []project.Keyframe{
		{ID: "k1", Time: 0, Property: "p", Value: 0.0, Easing: easing},
		{ID: "k2", Time: 1, Property: "p", Value: 1.0},
	}
}

func TestEasingParity(t *testing.T) {
	tests := []struct {
		easing string
		ref    func(float64) float64
	}{
		{"linear", func(u float64) float64 { return u }},
		{"", func(u float64) float64 { return u }},
		{"easeIn", refEaseIn},
		{"easeOut", refEaseOut},
		{"easeInOut", refEaseInOut},
		{"bounce", refBounce},
		{"elastic", refElastic},
		{"elasticOut", refElastic},
	}
	for _, tc := range tests {
		t.Run(tc.easing, func(t *testing.T) {
			expr := BuildAnimatedExpression(ratioTestKfs(tc.easing), "p", 0)
			for i := 0; i <= 20; i++ {
				u := float64(i) / 20
				got := evalAt(t, expr, u)
				want := tc.ref(u)
				if math.IsNaN(got) || math.IsInf(got, 0) {
					t.Fatalf("u=%v: expression yielded %v", u, got)
				}
				if math.Abs(got-want) > 1e-6 {
					t.Errorf("u=%v: got %v, want %v", u, got, want)
				}
			}
		})
	}
}

func TestBounceBoundaries(t *testing.T) {
	expr := BuildAnimatedExpression(ratioTestKfs("bounce"), "p", 0)
	// The classic easeOutBounce curve touches exactly 1.0 at each segment
	// boundary u = 1/2.75, 2/2.75, 2.5/2.75 and at u=1.
	for _, u := range []float64{1 / 2.75, 2 / 2.75, 2.5 / 2.75, 1} {
		got := evalAt(t, expr, u)
		if math.Abs(got-1) > 1e-9 {
			t.Errorf("bounce at u=%v: got %v, want 1", u, got)
		}
	}
	if v := evalAt(t, expr, 0); v != 0 {
		t.Errorf("bounce at u=0: got %v, want 0", v)
	}
	// Values in between are finite and within overshoot range (<=1).
	for i := 1; i < 20; i++ {
		u := float64(i) / 20
		v := evalAt(t, expr, u)
		if math.IsNaN(v) || v < -0.001 || v > 1.0001 {
			t.Errorf("bounce at u=%v out of range: %v", u, v)
		}
	}
}

func TestElasticFinite(t *testing.T) {
	expr := BuildAnimatedExpression(ratioTestKfs("elastic"), "p", 0)
	for i := 0; i <= 100; i++ {
		u := float64(i) / 100
		v := evalAt(t, expr, u)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("elastic at u=%v is not finite: %v", u, v)
		}
		// elastic out swings a little below 0 and above 1, but is bounded
		if v < -0.5 || v > 1.5 {
			t.Errorf("elastic at u=%v implausible: %v", u, v)
		}
	}
	if v := evalAt(t, expr, 0); v != 0 {
		t.Errorf("elastic at u=0: got %v, want 0", v)
	}
	if v := evalAt(t, expr, 1); math.Abs(v-1) > 1e-9 {
		t.Errorf("elastic at u=1: got %v, want 1", v)
	}
}

func TestElasticConstantPrecomputed(t *testing.T) {
	// 2*PI/3 must be folded on the Go side; PI is not available in geq.
	expr := buildEasingRatioExpr(0, 1, "elastic", "T")
	if strings.Contains(expr, "PI") {
		t.Errorf("elastic expr must not reference PI (undefined in geq): %s", expr)
	}
	if !strings.Contains(expr, "2.0943951023931953") {
		t.Errorf("elastic expr missing precomputed 2*PI/3 constant: %s", expr)
	}
}

func TestBuildAnimatedExpressionTUsesUppercaseVar(t *testing.T) {
	kfs := ratioTestKfs("bounce")
	expr := BuildAnimatedExpressionT(kfs, "p", 0)
	// No lowercase t identifier may survive (avoids matching "lt(", etc.).
	stripped := strings.ReplaceAll(expr, "lt(", "")
	if strings.Contains(stripped, "t") {
		t.Errorf("BuildAnimatedExpressionT must not contain lowercase t: %s", expr)
	}
	if v, err := parseExprAndEval(expr, map[string]float64{"T": 0.5}); err != nil {
		t.Fatalf("eval with T var failed: %v (expr %s)", err, expr)
	} else if math.Abs(v-refBounce(0.5)) > 1e-6 {
		t.Fatalf("at T=0.5 got %v, want %v", v, refBounce(0.5))
	}
}

func TestEqualKeyframeTimesDontDivideByZero(t *testing.T) {
	kfs := []project.Keyframe{
		{ID: "k1", Time: 1, Property: "x", Value: 5.0, Easing: "elastic"},
		{ID: "k2", Time: 1, Property: "x", Value: 15.0, Easing: "elastic"},
	}
	expr := BuildAnimatedExpression(kfs, "x", 0)
	if strings.Contains(expr, "/0") || strings.Contains(expr, "/-0") {
		t.Fatalf("division by zero in expr: %s", expr)
	}
	for _, time := range []float64{0.5, 1, 5} {
		v := evalAt(t, expr, time)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("t=%v: expr not finite: %v (%s)", time, v, expr)
		}
	}
}

func TestFloatToStrRoundtrip(t *testing.T) {
	// Constants emitted with %g must retain enough precision for easing
	// boundaries (FFmpeg prints %g with Go's shortest-representation).
	for _, f := range []float64{1 / 2.75, 2.625 / 2.75, 2 * math.Pi / 3} {
		s := FloatToStr(f)
		back, err := strconv.ParseFloat(s, 64)
		if err != nil || math.Abs(back-f) > 1e-12 {
			t.Errorf("FloatToStr(%v) = %q does not roundtrip", f, s)
		}
	}
}
