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

	var inputs []string
	for _, asset := range p.Assets {
		inputs = append(inputs, asset.Path)
	}
	args = append(args, inputs...)

	var filterParts []string

	videoIdx := 0
	audioIdx := 0

	for _, track := range p.Timeline.Tracks {
		for _, clip := range track.Clips {
			asset := findAsset(p.Assets, clip.AssetID)
			if asset == nil {
				continue
			}

			if track.Type == project.TrackVideo || track.Type == project.TrackText {
				filterParts = append(filterParts, fmt.Sprintf("[%d:v]trim=start=%g:end=%g,setpts=PTS-STARTPTS",
					videoIdx, clip.TrimStart, clip.TrimStart+clip.Duration))
				if tf := filters.BuildTransformFilter(clip.Transform); tf != "" {
					filterParts = append(filterParts, tf)
				}
				if cf := filters.BuildColorFilterChain(clip.Color); cf != "" {
					filterParts = append(filterParts, cf)
				}
				filterParts = append(filterParts, fmt.Sprintf("[v%d]", videoIdx))
				videoIdx++
			}

			if track.Type == project.TrackAudio || track.Type == project.TrackVideo {
				filterParts = append(filterParts, fmt.Sprintf("[%d:a]atrim=start=%g:end=%g,asetpts=PTS-STARTPTS",
					audioIdx, clip.TrimStart, clip.TrimStart+clip.Duration))
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
