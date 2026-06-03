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

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// Job carries all the data needed to render a single project.
type Job struct {
	ID         string
	Project    interface{}
	Settings   interface{}
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

	settings, ok := job.Settings.(map[string]interface{})
	if !ok {
		q.failJob(job, fmt.Errorf("invalid render settings payload"))
		return
	}
	outputPath, _ := settings["outputPath"].(string)
	if outputPath == "" {
		outputPath = job.OutputPath
	}
	if outputPath == "" {
		q.failJob(job, fmt.Errorf("no output path specified"))
		return
	}

	projectMap, _ := job.Project.(map[string]interface{})
	args := buildSimpleFFmpegArgs(projectMap, settings, outputPath)
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
	settings, _ := job.Settings.(map[string]interface{})
	if v, ok := settings["endTime"].(float64); ok && v > 0 {
		return v
	}
	projectMap, _ := job.Project.(map[string]interface{})
	if d, ok := projectMap["duration"].(float64); ok {
		return d
	}
	return 0
}

func (q *Queue) emitProgress(ev ProgressEvent) {
	if ev.JobID == "" {
		return
	}
	q.emit("render:progress", ev)
}

func buildSimpleFFmpegArgs(projectMap, settings map[string]interface{}, outputPath string) []string {
	if projectMap == nil {
		return nil
	}
	tracks, _ := projectMap["timeline"].(map[string]interface{})
	trackList, _ := tracks["tracks"].([]interface{})
	if len(trackList) == 0 {
		return nil
	}

	var args []string
	args = append(args, "ffmpeg", "-y")

	filterParts := []string{}
	inputIdx := 0
	for _, t := range trackList {
		track, _ := t.(map[string]interface{})
		trackType, _ := track["type"].(string)
		clips, _ := track["clips"].([]interface{})
		for _, c := range clips {
			clip, _ := c.(map[string]interface{})
			assetID, _ := clip["assetId"].(string)
			trimStart, _ := clip["trimStart"].(float64)
			duration, _ := clip["duration"].(float64)
			assetPath, _ := lookupAssetPath(projectMap, assetID)
			if assetPath == "" || duration <= 0 {
				continue
			}
			if trackType == "video" {
				if trimStart > 0 {
					args = append(args, "-ss", strconv.FormatFloat(trimStart, 'f', 3, 64))
				}
				args = append(args, "-i", assetPath, "-t", strconv.FormatFloat(duration, 'f', 3, 64))
				filterParts = append(filterParts, fmt.Sprintf("[%d:v]scale=1280:720:force_original_aspect_ratio=decrease,pad=1280:720:(ow-iw)/2:(oh-ih)/2,setsar=1,setpts=PTS-STARTPTS,fps=30[v%d]", inputIdx, inputIdx))
				inputIdx++
			}
		}
	}

	if len(filterParts) == 0 {
		return buildAudioOnlyArgs(projectMap, settings, outputPath)
	}

	concatIn := ""
	for i := 0; i < len(filterParts); i++ {
		concatIn += fmt.Sprintf("[v%d]", i)
	}
	filterParts = append(filterParts, fmt.Sprintf("%sconcat=n=%d:v=1:a=0[outv]", concatIn, len(filterParts)))

	args = append(args,
		"-filter_complex", strings.Join(filterParts, ";"),
		"-map", "[outv]",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
	)
	if ab, ok := settings["audioBitrate"].(string); ok && ab != "" {
		args = append(args, "-c:a", "aac", "-b:a", ab)
	}
	args = append(args, outputPath)
	return args
}

func buildAudioOnlyArgs(projectMap, settings map[string]interface{}, outputPath string) []string {
	var args []string
	args = append(args, "ffmpeg", "-y")

	filterParts := []string{}
	inputIdx := 0
	tracks, _ := projectMap["timeline"].(map[string]interface{})
	trackList, _ := tracks["tracks"].([]interface{})

	for _, t := range trackList {
		track, _ := t.(map[string]interface{})
		trackType, _ := track["type"].(string)
		if trackType != "audio" {
			continue
		}
		clips, _ := track["clips"].([]interface{})
		for _, c := range clips {
			clip, _ := c.(map[string]interface{})
			assetID, _ := clip["assetId"].(string)
			trimStart, _ := clip["trimStart"].(float64)
			duration, _ := clip["duration"].(float64)
			assetPath, _ := lookupAssetPath(projectMap, assetID)
			if assetPath == "" || duration <= 0 {
				continue
			}
			if trimStart > 0 {
				args = append(args, "-ss", strconv.FormatFloat(trimStart, 'f', 3, 64))
			}
			args = append(args, "-i", assetPath, "-t", strconv.FormatFloat(duration, 'f', 3, 64))
			filterParts = append(filterParts, fmt.Sprintf("[%d:a]asetpts=PTS-STARTPTS[a%d]", inputIdx, inputIdx))
			inputIdx++
		}
	}
	if len(filterParts) == 0 {
		return nil
	}
	mixIn := ""
	for i := 0; i < len(filterParts); i++ {
		mixIn += fmt.Sprintf("[a%d]", i)
	}
	filterParts = append(filterParts, fmt.Sprintf("%samix=inputs=%d:duration=longest[outa]", mixIn, len(filterParts)))

	args = append(args,
		"-filter_complex", strings.Join(filterParts, ";"),
		"-map", "[outa]",
		"-c:a", "aac",
		"-b:a", "192k",
		outputPath,
	)
	return args
}

func lookupAssetPath(projectMap map[string]interface{}, assetID string) (string, error) {
	assets, _ := projectMap["assets"].([]interface{})
	for _, a := range assets {
		am, _ := a.(map[string]interface{})
		if id, _ := am["id"].(string); id == assetID {
			if p, ok := am["path"].(string); ok {
				return p, nil
			}
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
