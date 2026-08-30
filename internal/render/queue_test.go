package render

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"Gocut/internal/project"
)

// mixedProject has one clip of every kind that contributes to the
// filtergraph, so an audio-only export has to prune several video branches.
func mixedProject(videoPath, audioPath, imagePath string) project.Project {
	return project.Project{
		Duration:   2,
		Resolution: project.Resolution{Width: 320, Height: 240},
		FPS:        30,
		Assets: []project.Asset{
			{ID: "vid", Path: videoPath, Type: project.AssetVideo, HasAudio: true, Duration: 2},
			{ID: "aud", Path: audioPath, Type: project.AssetAudio, HasAudio: true, Duration: 2},
			{ID: "img", Path: imagePath, Type: project.AssetImage, Duration: 2},
		},
		Timeline: project.Timeline{Tracks: []project.Track{
			{ID: "t1", Type: project.TrackVideo, Volume: 1, Clips: []project.Clip{
				{ID: "c1", AssetID: "vid", TrackID: "t1", Duration: 2, Speed: 1, Volume: 1, Opacity: 1},
			}},
			{ID: "t2", Type: project.TrackAudio, Volume: 1, Clips: []project.Clip{
				{ID: "c2", AssetID: "aud", TrackID: "t2", Duration: 2, Speed: 1, Volume: 1},
			}},
			{ID: "t3", Type: project.TrackImage, Volume: 1, Clips: []project.Clip{
				{ID: "c3", AssetID: "img", TrackID: "t3", Duration: 2, Speed: 1, Opacity: 1},
			}},
			{ID: "t4", Type: project.TrackText, Clips: []project.Clip{
				{ID: "c4", TrackID: "t4", Duration: 2, Speed: 1, Opacity: 1,
					TextProps: &project.TextProps{Text: "hello", FontSize: 24, Color: "#ffffff"}},
			}},
		}},
	}
}

func argValue(args []string, flag string) string {
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}

func TestBuildSimpleFFmpegArgsMP3IsAudioOnly(t *testing.T) {
	p := mixedProject("v.mp4", "a.mp3", "i.png")
	settings := project.RenderSettings{Format: "mp3", Codec: "mp3", AudioBitrate: "192k", EndTime: 2}

	args := buildSimpleFFmpegArgs(p, settings, "out.mp3", "ffmpeg")

	// The UI sends codec="mp3"; feeding that to -c:v is what made FFmpeg
	// abort with "Invalid encoder type 'mp3'".
	if hasFlag(args, "-c:v") {
		t.Errorf("mp3 export must not set a video encoder, got args: %v", args)
	}
	if got := argValue(args, "-c:a"); got != "libmp3lame" {
		t.Errorf("-c:a = %q, want libmp3lame", got)
	}
	if got := argValue(args, "-f"); got != "mp3" {
		t.Errorf("-f = %q, want mp3", got)
	}
	if got := argValue(args, "-b:a"); got != "192k" {
		t.Errorf("-b:a = %q, want 192k", got)
	}

	// Any labelled video pad left in the graph is unconsumed by -map and
	// makes FFmpeg exit with "has output N unconnected" / EINVAL.
	graph := argValue(args, "-filter_complex")
	for _, banned := range []string{"[basev]", "[v0]", "overlay=", "drawtext=", "scale="} {
		if strings.Contains(graph, banned) {
			t.Errorf("mp3 filtergraph must not contain %q, got: %s", banned, graph)
		}
	}
	if !strings.Contains(graph, "[outa]") {
		t.Errorf("mp3 filtergraph missing audio output, got: %s", graph)
	}
	if got := argValue(args, "-map"); got != "[outa]" {
		t.Errorf("-map = %q, want [outa]", got)
	}
}

func TestBuildSimpleFFmpegArgsMP3SkipsSilentInputs(t *testing.T) {
	p := mixedProject("v.mp4", "a.mp3", "i.png")
	p.Assets[0].HasAudio = false // video clip contributes nothing to an mp3
	settings := project.RenderSettings{Format: "mp3", Codec: "mp3", EndTime: 2}

	args := buildSimpleFFmpegArgs(p, settings, "out.mp3", "ffmpeg")

	joined := strings.Join(args, " ")
	for _, skipped := range []string{"v.mp4", "i.png"} {
		if strings.Contains(joined, skipped) {
			t.Errorf("mp3 export should not open %s, got args: %v", skipped, args)
		}
	}
	if !strings.Contains(joined, "a.mp3") {
		t.Errorf("mp3 export dropped the audio input, got args: %v", args)
	}
}

// TestMP3ExportRunsFFmpeg drives the real binary: the previous failure was
// an FFmpeg argument-validation error, so only an actual run proves it gone.
func TestMP3ExportRunsFFmpeg(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	dir := t.TempDir()
	videoPath := filepath.Join(dir, "v.mp4")
	audioPath := filepath.Join(dir, "a.wav")
	imagePath := filepath.Join(dir, "i.png")

	fixtures := [][]string{
		{"-y", "-f", "lavfi", "-i", "testsrc=size=320x240:rate=30:d=2",
			"-f", "lavfi", "-i", "sine=frequency=440:duration=2",
			"-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-shortest", videoPath},
		{"-y", "-f", "lavfi", "-i", "sine=frequency=880:duration=2", audioPath},
		{"-y", "-f", "lavfi", "-i", "color=c=red:size=320x240", "-frames:v", "1", imagePath},
	}
	for _, a := range fixtures {
		if out, err := exec.Command(ffmpegPath, a...).CombinedOutput(); err != nil {
			t.Fatalf("fixture generation failed: %v\n%s", err, out)
		}
	}

	p := mixedProject(videoPath, audioPath, imagePath)
	outputPath := filepath.Join(dir, "out.mp3")
	settings := project.RenderSettings{Format: "mp3", Codec: "mp3", AudioBitrate: "192k", EndTime: 2}

	args := buildSimpleFFmpegArgs(p, settings, outputPath, ffmpegPath)
	if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
		t.Fatalf("mp3 render failed: %v\nargs: %v\n%s", err, args, out)
	}

	probe, err := exec.Command("ffprobe", "-v", "error", "-show_entries",
		"stream=codec_type", "-of", "csv=p=0", outputPath).CombinedOutput()
	if err != nil {
		t.Fatalf("ffprobe failed: %v\n%s", err, probe)
	}
	streams := strings.TrimSpace(string(probe))
	if !strings.Contains(streams, "audio") {
		t.Errorf("expected an audio stream, ffprobe reported: %q", streams)
	}
	if strings.Contains(streams, "video") {
		t.Errorf("mp3 output must not carry a video stream, ffprobe reported: %q", streams)
	}
}

func TestBuildSimpleFFmpegArgsMP4KeepsVideo(t *testing.T) {
	p := mixedProject("v.mp4", "a.mp3", "i.png")
	settings := project.RenderSettings{Format: "mp4", Codec: "h264", EndTime: 2}

	args := buildSimpleFFmpegArgs(p, settings, "out.mp4", "ffmpeg")

	if got := argValue(args, "-c:v"); got != "h264" {
		t.Errorf("-c:v = %q, want h264", got)
	}
	graph := argValue(args, "-filter_complex")
	for _, want := range []string{"[basev]", "overlay=", "drawtext="} {
		if !strings.Contains(graph, want) {
			t.Errorf("mp4 filtergraph missing %q, got: %s", want, graph)
		}
	}
}

func TestBuildSimpleFFmpegArgsGIFDoesNotBuildAudioGraph(t *testing.T) {
	p := mixedProject("v.mp4", "a.mp3", "i.png")
	settings := project.RenderSettings{Format: "gif", Codec: "gif", EndTime: 2}

	args := buildSimpleFFmpegArgs(p, settings, "out.gif", "ffmpeg")
	graph := argValue(args, "-filter_complex")
	for _, banned := range []string{"[basea]", "amix=", "[outa]", "[va0]"} {
		if strings.Contains(graph, banned) {
			t.Errorf("GIF filtergraph must not contain %q, got: %s", banned, graph)
		}
	}
	if got := argValue(args, "-map"); got != "[gifout]" {
		t.Errorf("GIF map = %q, want [gifout]", got)
	}
	if !hasFlag(args, "-an") {
		t.Errorf("GIF export must disable audio, got args: %v", args)
	}
}

func TestBuildSimpleFFmpegArgsLoopUsesClipDuration(t *testing.T) {
	p := mixedProject("v.mp4", "a.mp3", "i.png")
	p.Duration = 10
	p.Timeline.Tracks[0].Clips[0].Loop = true
	p.Timeline.Tracks[0].Clips[0].Duration = 2
	settings := project.RenderSettings{Format: "mp4", Codec: "h264", EndTime: 10}

	args := buildSimpleFFmpegArgs(p, settings, "out.mp4", "ffmpeg")
	if got := argValue(args, "-stream_loop"); got != "-1" {
		t.Errorf("looped media must use -stream_loop -1, got args: %v", args)
	}
	if got := argValue(args, "-t"); got != "2.000" {
		t.Errorf("looped media input duration = %q, want clip duration 2.000", got)
	}
}

func TestGIFExportRunsFFmpegWithoutAudioGraph(t *testing.T) {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not on PATH")
	}
	dir := t.TempDir()
	videoPath := filepath.Join(dir, "v.mp4")
	outputPath := filepath.Join(dir, "out.gif")

	fixtureArgs := []string{
		"-y", "-f", "lavfi", "-i", "testsrc=size=160x120:rate=10:d=1",
		"-c:v", "libx264", "-pix_fmt", "yuv420p", videoPath,
	}
	if out, err := exec.Command(ffmpegPath, fixtureArgs...).CombinedOutput(); err != nil {
		t.Fatalf("GIF fixture generation failed: %v\n%s", err, out)
	}

	p := project.Project{
		Duration:   1,
		Resolution: project.Resolution{Width: 160, Height: 120},
		FPS:        10,
		Assets: []project.Asset{{
			ID: "vid", Path: videoPath, Type: project.AssetVideo,
			Width: 160, Height: 120, Duration: 1,
		}},
		Timeline: project.Timeline{Tracks: []project.Track{{
			ID: "video", Type: project.TrackVideo, Volume: 1,
			Clips: []project.Clip{{
				ID: "clip", AssetID: "vid", TrackID: "video",
				Duration: 1, Speed: 1, Volume: 1, Opacity: 1,
				Transform: project.Transform{ScaleX: 1, ScaleY: 1},
			}},
		}}},
	}
	settings := project.RenderSettings{Format: "gif", Codec: "gif", Width: 160, Height: 120, FPS: 10, EndTime: 1}

	args := buildSimpleFFmpegArgs(p, settings, outputPath, ffmpegPath)
	if out, err := exec.Command(args[0], args[1:]...).CombinedOutput(); err != nil {
		t.Fatalf("GIF render failed: %v\nargs: %v\n%s", err, args, out)
	}
}
