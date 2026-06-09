package render

import (
	"fmt"
	"strconv"
	"strings"

	"Gocut/internal/ffmpeg"
	"Gocut/internal/ffmpeg/filters"
	"Gocut/internal/project"
)

type Compositor struct {
	executor *ffmpeg.Executor
}

func NewCompositor(exe *ffmpeg.Executor) *Compositor {
	return &Compositor{executor: exe}
}

func (c *Compositor) BuildCommand(p project.Project, settings project.RenderSettings) ([]string, error) {
	var args []string

	args = append(args, "-y")

	if settings.StartTime > 0 {
		args = append(args, "-ss", ffmpeg.FloatToStr(settings.StartTime))
	}

	// Track asset -> input index so the filter graph can reference the
	// correct [N:v] stream for each clip. For still images we add
	// `-loop 1` so ffmpeg's image2 demuxer treats the file as a
	// frame-looped video stream that can be trimmed.
	assetToInput := make(map[string]int, len(p.Assets))
	for _, asset := range p.Assets {
		assetToInput[asset.ID] = len(args)
		if asset.Type == project.AssetImage {
			args = append(args, "-loop", "1")
		}
		args = append(args, "-i", asset.Path)
	}

	var filterParts []string

	videoIdx := 0
	audioIdx := 0

	for _, track := range p.Timeline.Tracks {
		for _, clip := range track.Clips {
			asset := findAsset(p.Assets, clip.AssetID)
			if asset == nil {
				continue
			}

			inputIdx := assetToInput[clip.AssetID]

			if track.Type == project.TrackVideo || track.Type == project.TrackText {
				if asset.Type == project.AssetImage {
					// Stills: hold the single frame for clip.Duration seconds.
					// We rely on the input already being declared with `-loop 1`.
					filterParts = append(filterParts, fmt.Sprintf(
						"[%d:v]trim=duration=%g,setpts=PTS-STARTPTS",
						inputIdx, clip.Duration))
				} else {
					filterParts = append(filterParts, fmt.Sprintf(
						"[%d:v]trim=start=%g:end=%g,setpts=PTS-STARTPTS",
						inputIdx, clip.TrimStart, clip.TrimStart+clip.Duration))
				}
				if tf := filters.BuildTransformFilter(clip); tf != "" {
					filterParts = append(filterParts, tf)
				}
				if cf := filters.BuildColorFilterChain(clip.Color); cf != "" {
					filterParts = append(filterParts, cf)
				}
				filterParts = append(filterParts, fmt.Sprintf("[v%d]", videoIdx))
				videoIdx++
			}

			if track.Type == project.TrackAudio || track.Type == project.TrackVideo {
				// Images have no audio stream; skip them to avoid
				// ffmpeg errors when a still is dragged in as video.
				if asset.Type == project.AssetImage {
					continue
				}
				filterParts = append(filterParts, fmt.Sprintf("[%d:a]atrim=start=%g:end=%g,asetpts=PTS-STARTPTS",
					inputIdx, clip.TrimStart, clip.TrimStart+clip.Duration))
				if af := filters.BuildAudioFilterChain(clip.Volume, false, false, 0.5, clip.Duration); af != "" {
					filterParts = append(filterParts, af)
				}
				filterParts = append(filterParts, fmt.Sprintf("[a%d]", audioIdx))
				audioIdx++
			}
		}
	}

	// Link video outputs
	if videoIdx > 1 {
		filterParts = append(filterParts, fmt.Sprintf("[v0][v1]concat=n=%d:v=1:a=0[vout]", videoIdx))
	} else if videoIdx == 1 {
		filterParts = append(filterParts, "[v0]copy[vout]")
	}

	// Mix audio
	if audioIdx > 1 {
		var amixInputs []string
		for i := 0; i < audioIdx; i++ {
			amixInputs = append(amixInputs, fmt.Sprintf("[a%d]", i))
		}
		filterParts = append(filterParts, strings.Join(amixInputs, "")+"amix=inputs="+strconv.Itoa(audioIdx)+"[aout]")
	} else if audioIdx == 1 {
		filterParts = append(filterParts, "[a0]copy[aout]")
	}

	filterComplex := strings.Join(filterParts, ";")
	if filterComplex != "" {
		args = append(args, "-filter_complex", filterComplex)
	}

	if videoIdx > 0 {
		args = append(args, "-map", "[vout]")
	}
	if audioIdx > 0 {
		args = append(args, "-map", "[aout]")
	}

	args = append(args, "-c:v", settings.Codec)
	args = append(args, "-crf", strconv.Itoa(settings.CRF))
	if settings.Preset != "" {
		args = append(args, "-preset", settings.Preset)
	}
	args = append(args, "-r", ffmpeg.FloatToStr(settings.FPS))
	args = append(args, "-s", fmt.Sprintf("%dx%d", settings.Width, settings.Height))

	if settings.AudioBitrate != "" {
		args = append(args, "-c:a", "aac", "-b:a", settings.AudioBitrate)
	}

	if settings.EndTime > 0 {
		args = append(args, "-t", ffmpeg.FloatToStr(settings.EndTime-settings.StartTime))
	}

	args = append(args, settings.OutputPath)

	return args, nil
}

func findAsset(assets []project.Asset, id string) *project.Asset {
	for i := range assets {
		if assets[i].ID == id {
			return &assets[i]
		}
	}
	return nil
}
