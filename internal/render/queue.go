package render

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"Gocut/internal/ffmpeg"
	"Gocut/internal/ffmpeg/filters"
	"Gocut/internal/project"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Job carries all the data needed to render a single project.
type Job struct {
	ID         string
	Project    project.Project
	Settings   project.RenderSettings
	OutputPath string
	Cmd        *exec.Cmd
	Cancel     context.CancelFunc
	Progress   float64
	Status     string
	Error      error
}

// ProgressEvent is the payload emitted to the frontend.
type ProgressEvent struct {
	JobID       string  `json:"jobId"`
	Percent     float64 `json:"percent"`
	CurrentTime float64 `json:"currentTime"`
	TotalTime   float64 `json:"totalTime"`
	FPS         float64 `json:"fps"`
	Status      string  `json:"status"`
	Error       string  `json:"error,omitempty"`
	OutputPath  string  `json:"outputPath,omitempty"`
}

// Queue is a single-worker FIFO render queue. The MVP intentionally runs
// one job at a time to avoid fighting over FFmpeg's CPU; multi-job queues
// land in v1.0.
type Queue struct {
	jobs   map[string]*Job
	mu     sync.Mutex
	nextID int
	ctx    context.Context
	cancel context.CancelFunc
	ready  chan *Job
	emit   func(event string, data interface{})
}

func NewQueue(ctx context.Context) *Queue {
	q := &Queue{
		jobs:  make(map[string]*Job),
		ready: make(chan *Job, 64),
	}
	q.ctx, q.cancel = context.WithCancel(ctx)
	q.emit = func(event string, data interface{}) {
		if ctx == nil {
			return
		}
		wailsruntime.EventsEmit(ctx, event, data)
	}
	go q.worker()
	return q
}

func (q *Queue) Enqueue(job *Job) string {
	q.mu.Lock()
	q.nextID++
	job.ID = fmt.Sprintf("job_%d_%d", time.Now().UnixNano(), q.nextID)
	q.jobs[job.ID] = job
	q.mu.Unlock()

	go func() {
		q.ready <- job
	}()
	return job.ID
}

func (q *Queue) Get(jobID string) *Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.jobs[jobID]
}

func (q *Queue) List() []*Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]*Job, 0, len(q.jobs))
	for _, j := range q.jobs {
		out = append(out, j)
	}
	return out
}

func (q *Queue) Cancel(jobID string) error {
	q.mu.Lock()
	job, ok := q.jobs[jobID]
	if !ok {
		q.mu.Unlock()
		return fmt.Errorf("job not found: %s", jobID)
	}
	if job.Cancel != nil {
		job.Cancel()
	}
	if job.Cmd != nil && job.Cmd.Process != nil {
		_ = job.Cmd.Process.Kill()
	}
	job.Status = "cancelled"
	q.mu.Unlock()

	q.emitProgress(ProgressEvent{
		JobID:  jobID,
		Status: "cancelled",
	})
	return nil
}

func (q *Queue) Stop() {
	q.cancel()
}

func (q *Queue) worker() {
	for {
		select {
		case <-q.ctx.Done():
			return
		case job := <-q.ready:
			q.runJob(job)
		}
	}
}

// runJob is the heart of the render pipeline.
func (q *Queue) runJob(job *Job) {
	q.mu.Lock()
	job.Status = "rendering"
	job.Progress = 0
	q.emitProgress(ProgressEvent{
		JobID:   job.ID,
		Percent: 0,
		Status:  "rendering",
	})
	q.mu.Unlock()

	settings := job.Settings
	outputPath := settings.OutputPath
	if outputPath == "" {
		outputPath = job.OutputPath
	}
	if outputPath == "" {
		q.failJob(job, fmt.Errorf("no output path specified"))
		return
	}

	p := job.Project
	args := buildSimpleFFmpegArgs(p, settings, outputPath)
	if len(args) == 0 {
		q.failJob(job, fmt.Errorf("no clips to render"))
		return
	}

	log.Printf("[render] ffmpeg args: %v", args)

	ctx, cancel := context.WithCancel(q.ctx)
	job.Cancel = cancel

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	ffmpeg.PrepareCmd(cmd)
	job.Cmd = cmd

	stderr, err := cmd.StderrPipe()
	if err != nil {
		q.failJob(job, fmt.Errorf("stderr pipe: %w", err))
		return
	}
	if err := cmd.Start(); err != nil {
		q.failJob(job, fmt.Errorf("start ffmpeg: %w", err))
		return
	}

	// Parse ffmpeg's stderr line-by-line for progress updates,
	// and collect tail lines so we can report errors.
	stderrLines := q.parseProgressAndCapture(job, stderr)

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return
		}
		// Include the last few stderr lines so the user sees the real FFmpeg error.
		tail := stderrTail(stderrLines, 8)
		log.Printf("[render] ffmpeg failed: %v\nstderr tail:\n%s", err, tail)
		q.failJob(job, fmt.Errorf("ffmpeg exited: %w\n%s", err, tail))
		return
	}

	q.mu.Lock()
	job.Status = "done"
	job.Progress = 100
	q.mu.Unlock()

	q.emitProgress(ProgressEvent{
		JobID:      job.ID,
		Percent:    100,
		Status:     "done",
		OutputPath: outputPath,
	})
}

func (q *Queue) failJob(job *Job, err error) {
	q.mu.Lock()
	job.Status = "error"
	job.Error = err
	q.mu.Unlock()
	q.emitProgress(ProgressEvent{
		JobID:  job.ID,
		Status: "error",
		Error:  err.Error(),
	})
}

// parseProgressAndCapture reads FFmpeg stderr, emits progress events, and
// returns all collected lines so the caller can report errors.
func (q *Queue) parseProgressAndCapture(job *Job, r io.ReadCloser) []string {
	defer r.Close()
	scanner := bufio.NewScanner(r)

	// FFmpeg uses \r (carriage return) for progress updates, not \n.
	// Default scanner splits on \n, meaning it blocks until end of job.
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := bytes.IndexAny(data, "\r\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines = append(lines, line)
		}
		// ffmpeg writes lines like:
		// frame=  120 fps= 30 q=29.0 size=    1024kB time=00:00:04.00 bitrate=2097.2kbits/s
		if !strings.Contains(line, "time=") {
			continue
		}
		cur := parseFFmpegTime(extractField(line, "time="))
		total := q.jobTotalDuration(job)
		if total > 0 {
			pct := (cur / total) * 100
			if pct > 100 {
				pct = 100
			}
			q.mu.Lock()
			job.Progress = pct
			q.mu.Unlock()
			q.emitProgress(ProgressEvent{
				JobID:       job.ID,
				Percent:     pct,
				CurrentTime: cur,
				TotalTime:   total,
				Status:      "rendering",
			})
		}
	}
	return lines
}

// stderrTail returns the last N non-empty lines from FFmpeg stderr.
func stderrTail(lines []string, n int) string {
	if len(lines) <= n {
		return strings.Join(lines, "\n")
	}
	return strings.Join(lines[len(lines)-n:], "\n")
}

func (q *Queue) jobTotalDuration(job *Job) float64 {
	duration := job.Settings.EndTime
	if duration <= 0 {
		duration = job.Project.Duration
	}
	if duration <= 0 {
		for _, track := range job.Project.Timeline.Tracks {
			for _, clip := range track.Clips {
				end := clip.StartTime + clip.Duration
				if end > duration {
					duration = end
				}
			}
		}
	}
	if duration <= 0 {
		duration = 10 // Fallback to avoid division by zero
	}
	return duration
}

func (q *Queue) emitProgress(ev ProgressEvent) {
	if ev.JobID == "" {
		return
	}
	q.emit("render:progress", ev)
}

func buildSimpleFFmpegArgs(p project.Project, settings project.RenderSettings, outputPath string) []string {
	if len(p.Timeline.Tracks) == 0 {
		return nil
	}

	var args []string
	args = append(args, "ffmpeg", "-y")

	// Calculate project duration based on the end time or max clip end
	duration := settings.EndTime
	if duration <= 0 {
		duration = p.Duration
	}
	if duration <= 0 {
		// Fallback to max clip end time
		for _, track := range p.Timeline.Tracks {
			for _, clip := range track.Clips {
				end := clip.StartTime + clip.Duration
				if end > duration {
					duration = end
				}
			}
		}
	}
	if duration <= 0 {
		duration = 10 // Fallback
	}

	filterParts := []string{}
	
	// Create base video and base audio
	filterParts = append(filterParts, fmt.Sprintf("color=c=black:s=1280x720:r=30:d=%.3f[basev]", duration))
	// Use atrim to limit anullsrc duration — more compatible across FFmpeg versions
	// than the 'd' parameter which some builds don't recognise on anullsrc.
	filterParts = append(filterParts, fmt.Sprintf("anullsrc=r=48000:cl=stereo,atrim=duration=%.3f[basea]", duration))

	inputIdx := 0
	audioLabels := []string{"[basea]"}

	type textOverlay struct {
		clip project.Clip
		tp   project.TextProps
	}
	var textClips []textOverlay

	type videoOverlay struct {
		label string
		clip  project.Clip
	}
	var videoOverlays []videoOverlay

	for _, track := range p.Timeline.Tracks {
		for _, clip := range track.Clips {
			if track.Type == project.TrackText {
				if clip.TextProps != nil {
					textClips = append(textClips, textOverlay{clip: clip, tp: *clip.TextProps})
				}
				continue
			}

			assetPath, err := lookupAssetPath(p, clip.AssetID)
			if err != nil || assetPath == "" {
				continue
			}

			ext := strings.ToLower(filepath.Ext(assetPath))
			isStaticImage := ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".bmp"

			if isStaticImage {
				args = append(args, "-loop", "1")
			} else if ext == ".gif" {
				args = append(args, "-ignore_loop", "0")
			}

			if clip.TrimStart > 0 {
				args = append(args, "-ss", strconv.FormatFloat(clip.TrimStart, 'f', 3, 64))
			}
			dur := clip.Duration
			if clip.Speed > 0 && clip.Speed != 1.0 {
				dur = dur * clip.Speed
			}
			args = append(args, "-t", strconv.FormatFloat(dur, 'f', 3, 64))
			args = append(args, "-i", assetPath)

			if track.Type == project.TrackVideo {
				label := fmt.Sprintf("[v%d]", inputIdx)
				var clipFilters []string

				clipFilters = append(clipFilters, "format=yuva420p")
				clipFilters = append(clipFilters, "scale=1280:720:force_original_aspect_ratio=decrease")
				clipFilters = append(clipFilters, "pad=1280:720:(ow-iw)/2:(oh-ih)/2")
				clipFilters = append(clipFilters, "setsar=1")

				if cf := filters.BuildColorFilterChain(clip.Color); cf != "" {
					clipFilters = append(clipFilters, cf)
				}
				if tf := ffmpeg.BuildTransformFilters(clip); tf != "" {
					clipFilters = append(clipFilters, tf)
				}
				if clip.Opacity > 0 && clip.Opacity < 1.0 {
					clipFilters = append(clipFilters, fmt.Sprintf("colorchannelmixer=aa=%g", clip.Opacity))
				}
				if clip.Transition != nil && clip.Transition.Type != "none" && clip.Transition.Duration > 0 {
					trans := clip.Transition
					D := trans.Duration
					switch trans.Type {
					case "fade", "dissolve":
						clipFilters = append(clipFilters, fmt.Sprintf("fade=t=in:st=0:d=%g:alpha=1", D))
					case "blur":
						clipFilters = append(clipFilters, fmt.Sprintf("boxblur=luma_radius='if(lt(t\\,%g)\\,20*(1-t/%g)\\,0)'", D, D))
					}
				}

				ptsExpr := "PTS-STARTPTS"
				if clip.Speed > 0 && clip.Speed != 1.0 {
					ptsExpr = fmt.Sprintf("(PTS-STARTPTS)/%g", clip.Speed)
				}
				clipFilters = append(clipFilters, fmt.Sprintf("setpts=%s+%.3f/TB", ptsExpr, clip.StartTime))

				filterParts = append(filterParts, fmt.Sprintf("[%d:v]%s%s", inputIdx, strings.Join(clipFilters, ","), label))
				videoOverlays = append(videoOverlays, videoOverlay{label: label, clip: clip})

				// Also process embedded audio from video clips
				aLabel := fmt.Sprintf("[va%d]", inputIdx)
				var aFilters []string
				aFilters = append(aFilters, "asetpts=PTS-STARTPTS")
				if clip.Speed > 0 && clip.Speed != 1.0 {
					aFilters = append(aFilters, fmt.Sprintf("atempo=%g", clip.Speed))
				}
				vol := clip.Volume
				if vol > 0 && vol != 1.0 {
					aFilters = append(aFilters, fmt.Sprintf("volume=%g", vol))
				}
				if track.Muted {
					aFilters = append(aFilters, "volume=0")
				}
				delayMs := int(clip.StartTime * 1000)
				aFilters = append(aFilters, fmt.Sprintf("adelay=%d|%d", delayMs, delayMs))
				filterParts = append(filterParts, fmt.Sprintf("[%d:a]%s%s", inputIdx, strings.Join(aFilters, ","), aLabel))
				audioLabels = append(audioLabels, aLabel)

			} else if track.Type == project.TrackImage || track.Type == project.TrackPIP {
				label := fmt.Sprintf("[v%d]", inputIdx)
				var clipFilters []string

				clipFilters = append(clipFilters, "format=yuva420p")
				clipFilters = append(clipFilters, "scale=1280:720:force_original_aspect_ratio=decrease")
				clipFilters = append(clipFilters, "pad=1280:720:(ow-iw)/2:(oh-ih)/2")
				clipFilters = append(clipFilters, "setsar=1")

				if cf := filters.BuildColorFilterChain(clip.Color); cf != "" {
					clipFilters = append(clipFilters, cf)
				}
				if tf := ffmpeg.BuildTransformFilters(clip); tf != "" {
					clipFilters = append(clipFilters, tf)
				}
				if clip.Opacity > 0 && clip.Opacity < 1.0 {
					clipFilters = append(clipFilters, fmt.Sprintf("colorchannelmixer=aa=%g", clip.Opacity))
				}
				if clip.Transition != nil && clip.Transition.Type != "none" && clip.Transition.Duration > 0 {
					trans := clip.Transition
					D := trans.Duration
					switch trans.Type {
					case "fade", "dissolve":
						clipFilters = append(clipFilters, fmt.Sprintf("fade=t=in:st=0:d=%g:alpha=1", D))
					case "blur":
						clipFilters = append(clipFilters, fmt.Sprintf("boxblur=luma_radius='if(lt(t\\,%g)\\,20*(1-t/%g)\\,0)'", D, D))
					}
				}

				ptsExpr := "PTS-STARTPTS"
				if clip.Speed > 0 && clip.Speed != 1.0 {
					ptsExpr = fmt.Sprintf("(PTS-STARTPTS)/%g", clip.Speed)
				}
				clipFilters = append(clipFilters, fmt.Sprintf("setpts=%s+%.3f/TB", ptsExpr, clip.StartTime))

				filterParts = append(filterParts, fmt.Sprintf("[%d:v]%s%s", inputIdx, strings.Join(clipFilters, ","), label))
				videoOverlays = append(videoOverlays, videoOverlay{label: label, clip: clip})

			} else if track.Type == project.TrackAudio {
				label := fmt.Sprintf("[a%d]", inputIdx)
				var clipFilters []string

				// Reset PTS first, then apply speed
				clipFilters = append(clipFilters, "asetpts=PTS-STARTPTS")
				if clip.Speed > 0 && clip.Speed != 1.0 {
					clipFilters = append(clipFilters, fmt.Sprintf("atempo=%g", clip.Speed))
				}

				fadeIn := clip.Transition != nil && clip.Transition.Type != "none" && clip.Transition.Duration > 0
				fadeDur := 0.0
				if fadeIn {
					fadeDur = clip.Transition.Duration
				}
				if af := ffmpeg.BuildAudioFilters(clip.Volume, fadeIn, false, fadeDur, clip.Duration); af != "" {
					clipFilters = append(clipFilters, af)
				}

				delayMs := int(clip.StartTime * 1000)
				clipFilters = append(clipFilters, fmt.Sprintf("adelay=%d|%d", delayMs, delayMs))

				filterParts = append(filterParts, fmt.Sprintf("[%d:a]%s%s", inputIdx, strings.Join(clipFilters, ","), label))
				audioLabels = append(audioLabels, label)
			}
			inputIdx++
		}
	}

	lastV := "[basev]"
	for i, vo := range videoOverlays {
		nextV := fmt.Sprintf("[ov%d]", i)
		xAnim := ffmpeg.BuildAnimatedExpression(vo.clip.Keyframes, "x", vo.clip.Transform.X)
		yAnim := ffmpeg.BuildAnimatedExpression(vo.clip.Keyframes, "y", vo.clip.Transform.Y)
		xExpr := fmt.Sprintf("(W-w)/2+%s", xAnim)
		yExpr := fmt.Sprintf("(H-h)/2+%s", yAnim)
		
		if vo.clip.Transition != nil && vo.clip.Transition.Type != "none" && vo.clip.Transition.Duration > 0 {
			trans := vo.clip.Transition
			start := vo.clip.StartTime
			dur := trans.Duration
			finalX := fmt.Sprintf("((W-w)/2+%.0f)", vo.clip.Transform.X)
			
			switch trans.Type {
			case "slideleft":
				xExpr = fmt.Sprintf("if(lt(t,%g),W-(W-%s)*(t-%g)/%g,%s)", start+dur, finalX, start, dur, finalX)
			case "slideright":
				xExpr = fmt.Sprintf("if(lt(t,%g),-w+(%s+w)*(t-%g)/%g,%s)", start+dur, finalX, start, dur, finalX)
			case "zoomin":
				// zoomin transition is implemented as scale in the clipFilters if needed, 
				// or we can simulate it by moving from center offset
			}
		}
		
		filterParts = append(filterParts, fmt.Sprintf("%s%soverlay=x='%s':y='%s':eof_action=pass%s", lastV, vo.label, xExpr, yExpr, nextV))
		lastV = nextV
	}

	for ti, tc := range textClips {
		text := escapeDrawtext(tc.tp.Text)
		fontSize := tc.tp.FontSize
		if fontSize <= 0 {
			fontSize = 48
		}
		fontColor := hexToFFmpeg(tc.tp.Color, "white")
		startT := tc.clip.StartTime
		endT := tc.clip.StartTime + tc.clip.Duration

		xExpr := "(w-text_w)/2"
		yExpr := "(h-text_h)/2"
		if tc.clip.Transform.X != 0 || tc.clip.Transform.Y != 0 {
			xExpr = fmt.Sprintf("(w-text_w)/2+%.0f", tc.clip.Transform.X)
			yExpr = fmt.Sprintf("(h-text_h)/2+%.0f", tc.clip.Transform.Y)
		}

		fontFile := ""
		if runtime.GOOS == "windows" {
			if tc.tp.Bold {
				fontFile = "C\\:/Windows/Fonts/arialbd.ttf"
			} else {
				fontFile = "C\\:/Windows/Fonts/arial.ttf"
			}
		} else if runtime.GOOS == "darwin" {
			if tc.tp.Bold {
				fontFile = "/System/Library/Fonts/HelveticaNeue-Bold.ttc"
			} else {
				fontFile = "/System/Library/Fonts/Helvetica.ttc"
			}
		} else {
			fontFile = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
		}

		drawtext := fmt.Sprintf("drawtext=text='%s':fontfile='%s':fontsize=%d:fontcolor=%s:x=%s:y=%s:enable='between(t\\,%.3f\\,%.3f)'",
			text, fontFile, fontSize, fontColor, xExpr, yExpr, startT, endT)

		if tc.tp.StrokeWidth > 0 {
			borderColor := hexToFFmpeg(tc.tp.StrokeColor, "black")
			drawtext += fmt.Sprintf(":borderw=%d:bordercolor=%s", tc.tp.StrokeWidth, borderColor)
		}

		if tc.tp.ShadowBlur > 0 || tc.tp.ShadowOffsetX != 0 || tc.tp.ShadowOffsetY != 0 {
			shadowColor := hexToFFmpeg(tc.tp.ShadowColor, "black@0.5")
			drawtext += fmt.Sprintf(":shadowcolor=%s:shadowx=%d:shadowy=%d", shadowColor, tc.tp.ShadowOffsetX, tc.tp.ShadowOffsetY)
		}

		nextLabel := fmt.Sprintf("[txt%d]", ti)
		filterParts = append(filterParts, fmt.Sprintf("%s%s%s", lastV, drawtext, nextLabel))
		lastV = nextLabel
	}

	// Mix all audio clips together
	lastA := "[outa]"
	if len(audioLabels) > 1 {
		mixIn := ""
		for _, l := range audioLabels {
			mixIn += l
		}
		filterParts = append(filterParts, fmt.Sprintf("%samix=inputs=%d:duration=first:dropout_transition=0[outa]", mixIn, len(audioLabels)))
	} else {
		// Just use base audio if no audio clips
		filterParts = append(filterParts, "[basea]anull[outa]")
	}

	args = append(args,
		"-filter_complex", strings.Join(filterParts, ";"),
		"-map", lastV,
		"-map", lastA,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
	)
	
	audioBitrate := settings.AudioBitrate
	if audioBitrate == "" {
		audioBitrate = "192k"
	}
	args = append(args, "-c:a", "aac", "-b:a", audioBitrate)
	
	args = append(args, outputPath)
	return args
}

// escapeDrawtext escapes special characters for FFmpeg drawtext filter.
func escapeDrawtext(s string) string {
	// FFmpeg drawtext needs single quotes escaped and colons/backslashes escaped
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "'\\''")
	s = strings.ReplaceAll(s, ":", "\\:")
	s = strings.ReplaceAll(s, "%", "%%")
	return s
}

// hexToFFmpeg converts a CSS hex color like "#ff0000" or "rgba(255,0,0,1)" to FFmpeg color format.
func hexToFFmpeg(color, fallback string) string {
	if color == "" {
		return fallback
	}
	// Strip # prefix — FFmpeg accepts hex without #, or named colors
	if strings.HasPrefix(color, "#") {
		return "0x" + strings.TrimPrefix(color, "#")
	}
	// For rgba() or other CSS formats, just return fallback
	if strings.HasPrefix(color, "rgb") {
		return fallback
	}
	return color
}

func lookupAssetPath(p project.Project, assetID string) (string, error) {
	for _, a := range p.Assets {
		if a.ID == assetID {
			return a.Path, nil
		}
	}
	return "", fmt.Errorf("asset not found: %s", assetID)
}

func extractField(line, key string) string {
	idx := strings.Index(line, key)
	if idx < 0 {
		return ""
	}
	rest := line[idx+len(key):]
	end := strings.IndexAny(rest, " \t\n")
	if end < 0 {
		return rest
	}
	return rest[:end]
}

func parseFFmpegTime(s string) float64 {
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.ParseFloat(parts[0], 64)
	m, _ := strconv.ParseFloat(parts[1], 64)
	sec, _ := strconv.ParseFloat(parts[2], 64)
	return h*3600 + m*60 + sec
}
