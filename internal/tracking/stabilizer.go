package tracking

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

type Stabilizer struct {
	FFmpegPath string
	TempDir    string
}

func NewStabilizer(ffmpegPath string) *Stabilizer {
	return &Stabilizer{
		FFmpegPath: ffmpegPath,
		TempDir:    os.TempDir(),
	}
}

func (s *Stabilizer) ApplyStabilization(assetPath string, trackingData *TrackingData, outputPath string, onProgress func(float64)) error {
	if trackingData.TransformFile == "" {
		return fmt.Errorf("no transform file available for stabilization")
	}

	smoothing := 10 // Default smoothing factor
	vf := fmt.Sprintf("vidstabtransform=input='%s':smoothing=%d:optzoom=0",
		trackingData.TransformFile, smoothing)

	args := []string{
		"-i", assetPath,
		"-vf", vf,
		"-c:v", "libx264",
		"-crf", "18",
		"-preset", "fast",
		"-y",
		outputPath,
	}

	cmd := exec.Command(s.FFmpegPath, args...)
	stderr, _ := cmd.StderrPipe()

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Parse progress
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		// Parse time= for progress
		if onProgress != nil {
			// Simple progress estimate
			onProgress(0.5)
		}
	}

	return <-done
}
