package app

import (
	"os/exec"
	"path/filepath"
	"testing"

	"Gocut/internal/project"
)

// TestPreviewChromaKey verifies that buildPreviewArgs emits a valid ffmpeg
// graph that applies the clip's color chain (incl. chromakey) and composites
// over black, turning a green frame's center pixel black.
func TestPreviewChromaKey(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not on PATH")
	}

	dir := t.TempDir()
	green := filepath.Join(dir, "green.png")
	gen := exec.Command(ffmpegPath, "-f", "lavfi", "-i", "color=c=0x00ff00:s=200x200:d=0.1", "-frames:v", "1", green)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("generate green frame: %v\n%s", err, out)
	}

	p := project.Project{
		Resolution: project.Resolution{Width: 200, Height: 200},
		FPS:        30,
		Assets:     []project.Asset{{ID: "a1", Path: green, Type: project.AssetVideo, Width: 200, Height: 200}},
		Timeline: project.Timeline{
			Tracks: []project.Track{{
				Type: project.TrackVideo,
				Clips: []project.Clip{{
					ID: "c1", AssetID: "a1", StartTime: 0, Duration: 1,
					Color: project.ColorGrade{
						ChromaKeyColor:      "#00ff00",
						ChromaKeySimilarity: 0.1,
						ChromaKeyBlend:      0,
					},
				}},
			}},
		},
	}

	layers := visualLayersAtTime(p, 0)
	if len(layers) != 1 {
		t.Fatalf("expected 1 layer, got %d", len(layers))
	}
	args, err := buildPreviewArgs(p, layers, 200, 200, 0)
	if err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(ffmpegPath, args...)
	jpg, err := cmd.Output()
	if err != nil {
		t.Fatalf("preview run failed: %v\n%s", err, cmd.Stderr)
	}
	if len(jpg) == 0 {
		t.Fatal("empty preview output")
	}

	// Sample the center pixel of the produced frame; it must be black
	// because chroma key removed the green. Decode the JPEG to raw rgb24
	// and read the pixel at (100,100) of the 200x200 frame.
	const w, h = 200, 200
	probe := exec.Command(ffmpegPath, "-i", "-", "-vf", "format=rgb24", "-f", "rawvideo", "-")
	stdin, err := probe.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		defer stdin.Close()
		stdin.Write(jpg)
	}()
	raw, err := probe.Output()
	if err != nil {
		t.Fatalf("probe failed: %v\n%s", err, probe.Stderr)
	}
	need := w * h * 3
	if len(raw) < need {
		t.Fatalf("raw frame too small: %d bytes, want %d", len(raw), need)
	}
	idx := (100*w + 100) * 3
	pix := raw[idx : idx+3]
	if pix[0] != 0 || pix[1] != 0 || pix[2] != 0 {
		t.Fatalf("expected black center pixel after chroma key, got %d %d %d", pix[0], pix[1], pix[2])
	}
}
