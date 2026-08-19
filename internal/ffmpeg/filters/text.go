package filters

import (
	"fmt"
	"strings"
)

// Animated text export route
// =============================
//
// Animated text clips cannot use the plain "drawtext into the base stream"
// pattern that static text uses: fade / slide / bounce / zoom need either a
// per-frame overlay position, an alpha pad, a per-frame scale or an alpha
// mask pass. So each animated clip is rendered on its own full-frame
// transparent layer and overlaid back with an enable window:
//
//	color=c=black@0:s=WxH:r=FPS:d=CLIPDUR,format=yuva420p,
//	drawtext(<full styling, centred on layer> + Transform offset)
//	[,fade alpha=1 | scale eval=frame | geq alpha-wipe]
//	setpts=PTS+START/TB                                  [tfN]
//	[prevV][tfN]overlay=x='..':y='..':enable='between(t\,START\,END)':eof_action=pass [txtN]
//
// Every construct was tested empirically against the ffmpeg on PATH
// (2025-12-18-git-78c75d546a-full_build, Windows) with `-f null -` before
// being used here. Gotchas discovered the hard way:
//
//  1. drawtext with no explicit fontfile segfaults on this Windows build
//     ("Fontconfig error: Cannot load default config file"); an explicit
//     fontfile is mandatory. This is why both the static and animated paths
//     always pass fontfile=
//  2. drawtext alpha='expr' works, but the layer approach covers every
//     preset uniformly, so it is used for all animated presets.
//  3. fade=alpha=1 on a yuva420p layer fades ONLY the alpha channel. OK.
//  4. scale=eval=frame ramps work, but dimensions must stay even — wrap
//     results in floor(x/2)*2.
//  5. overlay x/y accept animated expressions using the MAIN input's `t`
//     (same convention as the video overlay loop in render/queue.go).
//  6. geq alpha wipe: on this build the plane accessors are lum/cb/cr/alpha
//     — a(X,Y) fails with "Unknown function" — and the time variable is T
//     (uppercase); lowercase t is undefined at graph init time. The mask is
//     a multiplicative factor 0..1 applied to the existing alpha plane.
//     Validated expression:
//       geq=lum='lum(X,Y)':cb='cb(X,Y)':cr='cr(X,Y)':a='alpha(X,Y)*lt(X,W*min(1,T/D))'
//  7. typewriter (per-char reveal) is not expressible with drawtext: text_w
//     is the total line width and there is no per-char geometry at graph
//     init time. The export DEGRADES it to fade_in; the CSS preview keeps
//     the real per-char reveal.

// AnimatedTextSpec carries the per-clip inputs for one animated text layer.
// Text is RAW — BuildAnimatedTextLayers escapes it internally with the same
// rule as render/queue.go's escapeDrawtext. FontColor / BorderColor /
// ShadowColor must already be converted by hexToFFmpeg.
type AnimatedTextSpec struct {
	Text        string // raw text, escaped internally
	FontFile    string
	FontSize    int
	FontColor   string
	X, Y        string // drawtext position expressions (centred + offset)
	StrokeWidth int
	BorderColor string
	HasShadow   bool
	ShadowColor string
	ShadowX     int
	ShadowY     int
	Start       float64 // clip start on the global timeline
	End         float64 // Start + ClipDur
	ClipDur     float64
	W, H        int // project size
	FPS         float64
	InLabel     string  // current chain label, e.g. "[ov0]" / "[txt1]"
	LayerName   string  // layer label, e.g. "[tf0]"
	OutName     string  // overlay output label, e.g. "[txt0]"
	Anim        string  // resolved preset id (never "" here)
	Dur         float64 // resolved animation duration in seconds
}

// ResolveTextAnim normalises the preset id + duration and applies the export
// degradation rules. A resolved anim of "" means: static text — the caller
// must use the legacy inline drawtext path (byte-identical regression path).
func ResolveTextAnim(anim string, dur float64) (string, float64) {
	switch anim {
	case "", "none":
		return "", 0
	case "typewriter":
		// Not reproducible with drawtext (see file header, note 7).
		// Degrade to fade_in on export; preview keeps true typewriter.
		anim = "fade_in"
	}
	if dur <= 0 {
		dur = 0.5
	}
	return anim, dur
}

// escapeDrawText escapes text for drawtext=, byte-for-byte the same rule
// as render/queue.go's private escapeDrawtext. Keep in sync.
func escapeDrawText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "'\\''")
	s = strings.ReplaceAll(s, ":", "\\:")
	s = strings.ReplaceAll(s, "%", "%%")
	return s
}

// BuildAnimatedTextLayers returns the two filter_complex entries that render
// one animated text clip: the layer chain first, then the overlay back onto
// InLabel. The caller appends them in order and continues the chain from
// spec.OutName.
func BuildAnimatedTextLayers(s AnimatedTextSpec) []string {
	fontSize := s.FontSize
	if fontSize <= 0 {
		fontSize = 48
	}
	color := s.FontColor
	if color == "" {
		color = "white"
	}

	drawtext := fmt.Sprintf("drawtext=text='%s':fontfile='%s':fontsize=%d:fontcolor=%s:x=%s:y=%s",
		escapeDrawText(s.Text), s.FontFile, fontSize, color, s.X, s.Y)

	if s.StrokeWidth > 0 {
		border := s.BorderColor
		if border == "" {
			border = "black"
		}
		drawtext += fmt.Sprintf(":borderw=%d:bordercolor=%s", s.StrokeWidth, border)
	}
	if s.HasShadow {
		sh := s.ShadowColor
		if sh == "" {
			sh = "black@0.5"
		}
		drawtext += fmt.Sprintf(":shadowcolor=%s:shadowx=%d:shadowy=%d", sh, s.ShadowX, s.ShadowY)
	}

	// drawtext centres on the full-frame layer, so overlay x/y keep the
	// layer aligned with the main frame (0,0) unless the preset animates
	// the overlay position itself (slide_*, bounce).
	parts := []string{
		fmt.Sprintf("color=c=black@0:s=%dx%d:r=%.0f:d=%.3f", s.W, s.H, s.FPS, s.ClipDur),
		"format=yuva420p",
		drawtext,
	}
	if preset := presetLayerFilter(s.Anim, s.Dur, s.ClipDur); preset != "" {
		parts = append(parts, preset)
	}
	// setpts LAST: preset effects (fade/scale/geq) read layer-local time
	// starting at 0; the shift aligns the layer with the global timeline.
	parts = append(parts, fmt.Sprintf("setpts=PTS+%.3f/TB", s.Start))

	layer := fmt.Sprintf("%s%s", strings.Join(parts, ","), s.LayerName)

	ox, oy := overlayXYExpr(s)
	overlay := fmt.Sprintf("%s%soverlay=x='%s':y='%s':enable='between(t\\,%.3f\\,%.3f)':eof_action=pass%s",
		s.InLabel, s.LayerName, ox, oy, s.Start, s.End, s.OutName)

	return []string{layer, overlay}
}

// overlayXYExpr returns the overlay x/y expressions for the animated layer.
// Presets that don't move return 0,0 — the text is already centred on the
// full-frame layer.
func overlayXYExpr(s AnimatedTextSpec) (x, y string) {
	// cubic easeOut over Dur after Start: 1 -> 0
	easeIn := fmt.Sprintf("pow(max(0\\,(1-(t-%.3f)/%g))\\,3)", s.Start, s.Dur)
	switch s.Anim {
	case "slide_left": // enter from the right edge
		return fmt.Sprintf("W*%s", easeIn), "0"
	case "slide_right": // enter from the left edge
		return fmt.Sprintf("-w*%s", easeIn), "0"
	case "slide_top": // enter from the top edge
		return "0", fmt.Sprintf("-h*%s", easeIn)
	case "slide_bottom": // enter from the bottom edge
		return "0", fmt.Sprintf("H*%s", easeIn)
	case "bounce":
		// decaying hop: |sin(u·PI·3)|·(1-u)·40% of frame height
		u := fmt.Sprintf("min(1\\,(t-%.3f)/%g)", s.Start, s.Dur)
		hop := fmt.Sprintf("abs(sin(%s*PI*3))*(1-%s)*0.4*H", u, u)
		return "0", fmt.Sprintf("-%s", hop)
	default:
		return "0", "0"
	}
}

// presetLayerFilter returns the per-preset layer effect (fade / scale /
// geq). Presets animated purely via overlay position return "".
func presetLayerFilter(anim string, dur, clipDur float64) string {
	switch anim {
	case "fade_in":
		return fmt.Sprintf("fade=t=in:st=0:d=%g:alpha=1", dur)
	case "fade_out":
		st := clipDur - dur
		if st < 0 {
			st = 0
		}
		return fmt.Sprintf("fade=t=out:st=%g:d=%g:alpha=1", st, dur)
	case "zoom_in":
		// k runs 3 -> 1 with cubic easeOut; a short alpha fade hides the
		// hard pop of the oversized first frame.
		k := fmt.Sprintf("(3-2*pow(min(t\\/%g\\,1)\\,3))", dur)
		return fmt.Sprintf(
			"fade=t=in:st=0:d=%g:alpha=1,scale=w='floor(iw*%s/2)*2':h='floor(ih*%s/2)*2':eval=frame",
			dur*0.6, k, k)
	case "pop":
		// two-stage scale: grow to 1.15 over the first 70% of Dur, settle
		// to 1.0 over the rest — matches the preview overshoot.
		p1 := dur * 0.7
		k := fmt.Sprintf("if(lt(t\\,%g)\\,(t/%g)*1.15\\,1.15-((t-%g)/%g)*0.15)", p1, p1, p1, dur-p1)
		return fmt.Sprintf(
			"fade=t=in:st=0:d=%g:alpha=1,scale=w='floor(iw*%s/2)*2':h='floor(ih*%s/2)*2':eval=frame",
			dur/3, k, k)
	case "wipe":
		// empirically validated: alpha() accessor + T time var (see header).
		return fmt.Sprintf(
			"geq=lum='lum(X,Y)':cb='cb(X,Y)':cr='cr(X,Y)':a='alpha(X,Y)*lt(X,W*min(1,T/%g))'", dur)
	default:
		// bounce / slide_* are driven by overlayXYExpr.
		return ""
	}
}
