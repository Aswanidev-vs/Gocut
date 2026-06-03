package render

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

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

	ctx, cancel := context.WithCancel(q.ctx)
	job.Cancel = cancel

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
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

	// Parse ffmpeg's stderr line-by-line for progress updates.
	go q.parseProgress(job, stderr)

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return
		}
		q.failJob(job, fmt.Errorf("ffmpeg exited: %w", err))
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

func (q *Queue) parseProgress(job *Job, r io.ReadCloser) {
	defer r.Close()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
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
}

func (q *Queue) jobTotalDuration(job *Job) float64 {
	if job.Settings.EndTime > 0 {
		return job.Settings.EndTime
	}
	if job.Project.Duration > 0 {
		return job.Project.Duration
	}
	return 0
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
	filterParts = append(filterParts, fmt.Sprintf("anullsrc=r=48000:cl=stereo:d=%.3f[basea]", duration))

	inputIdx := 0
	videoLabels := []string{"[basev]"}
	audioLabels := []string{"[basea]"}

	for _, track := range p.Timeline.Tracks {
		for _, clip := range track.Clips {
			assetPath, err := lookupAssetPath(p, clip.AssetID)
			if err != nil || assetPath == "" || clip.Duration <= 0 {
				continue
			}

			// Add input options BEFORE -i for accurate seeking and duration trimming
			if clip.TrimStart > 0 {
				args = append(args, "-ss", strconv.FormatFloat(clip.TrimStart, 'f', 3, 64))
			}
			args = append(args, "-t", strconv.FormatFloat(clip.Duration, 'f', 3, 64))
			args = append(args, "-i", assetPath)

			if track.Type == project.TrackVideo {
				label := fmt.Sprintf("[v%d]", inputIdx)
				filterParts = append(filterParts, fmt.Sprintf("[%d:v]scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2,setsar=1,setpts=PTS-STARTPTS+%.3f/TB%s", inputIdx, clip.StartTime, label))
				videoLabels = append(videoLabels, label)
			} else if track.Type == project.TrackAudio {
				label := fmt.Sprintf("[a%d]", inputIdx)
				delayMs := int(clip.StartTime * 1000)
				filterParts = append(filterParts, fmt.Sprintf("[%d:a]asetpts=PTS-STARTPTS,adelay=%d|%d%s", inputIdx, delayMs, delayMs, label))
				audioLabels = append(audioLabels, label)
			}
			inputIdx++
		}
	}

	// Overlay all video clips onto the base video sequentially
	lastV := videoLabels[0]
	for i := 1; i < len(videoLabels); i++ {
		nextV := fmt.Sprintf("[ov%d]", i)
		filterParts = append(filterParts, fmt.Sprintf("%s%soverlay=eof_action=pass%s", lastV, videoLabels[i], nextV))
		lastV = nextV
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
