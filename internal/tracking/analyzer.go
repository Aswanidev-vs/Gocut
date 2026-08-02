package tracking

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Analyzer struct {
	FFmpegPath string
	TempDir    string
}

func NewAnalyzer(ffmpegPath string) *Analyzer {
	return &Analyzer{
		FFmpegPath: ffmpegPath,
		TempDir:    os.TempDir(),
	}
}

func (a *Analyzer) AnalyzeMotion(settings TrackingSettings, assetPath string, onProgress func(float64)) (*TrackingData, error) {
	switch settings.Method {
	case TrackStabilize:
		return a.analyzeStabilize(settings, assetPath, onProgress)
	case TrackPoint:
		return a.analyzePointTrack(settings, assetPath, onProgress)
	default:
		return nil, fmt.Errorf("unknown tracking method: %s", settings.Method)
	}
}

func (a *Analyzer) analyzeStabilize(settings TrackingSettings, assetPath string, onProgress func(float64)) (*TrackingData, error) {
	trfFile := filepath.Join(a.TempDir, fmt.Sprintf("gocut_stab_%d.trf", os.Getpid()))

	shakiness := settings.Shaking + 3 // FFmpeg range: 1-10, we map 0-2 to 3-5
	accuracy := settings.Accuracy + 1  // FFmpeg range: 1-4, we map 0-2 to 1-3

	vf := fmt.Sprintf("vidstabdetect=shakiness=%d:accuracy=%d:result='%s'", shakiness, accuracy, trfFile)

	args := []string{
		"-ss", fmt.Sprintf("%.3f", settings.StartTime),
		"-t", fmt.Sprintf("%.3f", settings.Duration),
		"-i", assetPath,
		"-vf", vf,
		"-f", "null",
		"-y",
		"/dev/null",
	}

	cmd := exec.Command(a.FFmpegPath, args...)
	stderr, _ := cmd.StderrPipe()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Parse progress from stderr
	re := regexp.MustCompile(`frame=\s*(\d+)\s+fps=\s*[\d.]+`)
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); matches != nil {
			frame, _ := strconv.Atoi(matches[1])
			if onProgress != nil && settings.Duration > 0 {
				// Rough progress estimate
				onProgress(float64(frame) / (settings.Duration * 30))
			}
		}
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("stabilization analysis failed: %w", err)
	}

	// Parse the .trf file to extract motion vectors
	points, err := parseTRFFile(trfFile, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to parse transform file: %w", err)
	}

	return &TrackingData{
		AssetID:       settings.AssetID,
		Method:        TrackStabilize,
		Points:        points,
		TransformFile: trfFile,
		Confidence:    0.9,
		FrameCount:    len(points),
	}, nil
}

func (a *Analyzer) analyzePointTrack(settings TrackingSettings, assetPath string, onProgress func(float64)) (*TrackingData, error) {
	// Extract frames to temp directory
	framesDir := filepath.Join(a.TempDir, fmt.Sprintf("gocut_frames_%d", os.Getpid()))
	os.MkdirAll(framesDir, 0755)
	defer os.RemoveAll(framesDir)

	// Extract frames
	extractArgs := []string{
		"-ss", fmt.Sprintf("%.3f", settings.StartTime),
		"-t", fmt.Sprintf("%.3f", settings.Duration),
		"-i", assetPath,
		"-vf", fmt.Sprintf("fps=30"),
		"-q:v", "2",
		filepath.Join(framesDir, "frame_%04d.jpg"),
	}

	cmd := exec.Command(a.FFmpegPath, extractArgs...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("frame extraction failed: %w", err)
	}

	// Read extracted frames
	entries, err := os.ReadDir(framesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read frames directory: %w", err)
	}

	var frameFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".jpg") {
			frameFiles = append(frameFiles, filepath.Join(framesDir, e.Name()))
		}
	}

	// Simple template matching using FFmpeg geq for each frame
	points := make([]TrackedPoint, 0, len(frameFiles))
	for i, frameFile := range frameFiles {
		if onProgress != nil {
			onProgress(float64(i) / float64(len(frameFiles)))
		}

		point, err := a.trackPointInFrame(frameFile, settings)
		if err != nil {
			continue // Skip frames where tracking fails
		}
		point.Frame = i
		point.Time = float64(i) / 30.0
		points = append(points, *point)
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("no points could be tracked")
	}

	return &TrackingData{
		AssetID:    settings.AssetID,
		Method:     TrackPoint,
		Points:     points,
		Confidence: 0.8,
		FrameCount: len(points),
	}, nil
}

func (a *Analyzer) trackPointInFrame(frameFile string, settings TrackingSettings) (*TrackedPoint, error) {
	// For now, return the center of the tracked region as a placeholder
	// A real implementation would use template matching or feature detection
	return &TrackedPoint{
		X: float64(settings.RegionX + settings.RegionW/2),
		Y: float64(settings.RegionY + settings.RegionH/2),
	}, nil
}

func parseTRFFile(trfPath string, settings TrackingSettings) ([]TrackedPoint, error) {
	file, err := os.Open(trfPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var points []TrackedPoint
	scanner := bufio.NewScanner(file)
	frame := 0
	fps := 30.0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// vidstab .trf format: frame dx dy da
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			dx, _ := strconv.ParseFloat(fields[0], 64)
			dy, _ := strconv.ParseFloat(fields[1], 64)

			points = append(points, TrackedPoint{
				Frame: frame,
				Time:  float64(frame) / fps,
				X:     -dx, // Invert for stabilization
				Y:     -dy,
			})
			frame++
		}
	}

	return points, nil
}
