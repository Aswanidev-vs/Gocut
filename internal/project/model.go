package project

import (
	"encoding/json"
	"time"
)

type TrackType string

const (
	TrackVideo   TrackType = "video"
	TrackAudio   TrackType = "audio"
	TrackImage   TrackType = "image"
	TrackPIP     TrackType = "pip"
	TrackText    TrackType = "text"
	TrackSticker TrackType = "sticker"
	TrackFX      TrackType = "fx"
)

type Resolution struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// ProjectSettings is the payload the frontend sends to NewProject.
// It carries both the project identity (name/aspect/resolution/fps) and
// the editor settings (background color, autosave).
type ProjectSettings struct {
	Name             string      `json:"name"`
	AspectRatio      string      `json:"aspectRatio"`
	Resolution       *Resolution `json:"resolution,omitempty"`
	FPS              float64     `json:"fps"`
	BackgroundColor  string      `json:"backgroundColor"`
	AutoSave         bool        `json:"autoSave"`
	AutoSaveInterval int         `json:"autoSaveIntervalSeconds"`
}

type Project struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
	Duration            float64         `json:"duration"`
	AspectRatio         string          `json:"aspectRatio"`
	Resolution          Resolution      `json:"resolution"`
	FPS                 float64         `json:"fps"`
	Timeline            Timeline        `json:"timeline"`
	Assets              []Asset         `json:"assets"`
	DesignGraph         json.RawMessage `json:"designGraph,omitempty"`
	Settings            ProjectSettings `json:"settings"`
	FilePath            string          `json:"filePath,omitempty"`
	CustomSaveDirectory string          `json:"customSaveDirectory,omitempty"`
}

type Timeline struct {
	Tracks   []Track `json:"tracks"`
	Duration float64 `json:"duration"`
}

type Track struct {
	ID     string    `json:"id"`
	Type   TrackType `json:"type"`
	Clips  []Clip    `json:"clips"`
	Muted  bool      `json:"muted"`
	Locked bool      `json:"locked"`
	Volume float64   `json:"volume"`
}

type Transform struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	ScaleX   float64 `json:"scaleX"`
	ScaleY   float64 `json:"scaleY"`
	Rotation float64 `json:"rotation"`
	FlipH    bool    `json:"flipH"`
	FlipV    bool    `json:"flipV"`
	CropX    float64 `json:"cropX"`
	CropY    float64 `json:"cropY"`
	CropW    float64 `json:"cropW"`
	CropH    float64 `json:"cropH"`
}

type ColorGrade struct {
	Brightness          int     `json:"brightness"`
	Contrast            int     `json:"contrast"`
	Saturation          int     `json:"saturation"`
	Hue                 int     `json:"hue"`
	Sharpness           int     `json:"sharpness"`
	Vignette            int     `json:"vignette"`
	Grain               int     `json:"grain"`
	Blur                int     `json:"blur"`
	Tint                int     `json:"tint"`
	Temp                int     `json:"temp"`
	Highlights          int     `json:"highlights"`
	Shadows             int     `json:"shadows"`
	LiftR               int     `json:"liftR"`
	LiftG               int     `json:"liftG"`
	LiftB               int     `json:"liftB"`
	GammaR              int     `json:"gammaR"`
	GammaG              int     `json:"gammaG"`
	GammaB              int     `json:"gammaB"`
	GainR               int     `json:"gainR"`
	GainG               int     `json:"gainG"`
	GainB               int     `json:"gainB"`
	Curves              string  `json:"curves"`
	ChromaKeyColor      string  `json:"chromaKeyColor"`
	ChromaKeySimilarity float64 `json:"chromaKeySimilarity"`
	ChromaKeyBlend      float64 `json:"chromaKeyBlend"`
	LutPath             string  `json:"lutPath,omitempty"`
}

type TextProps struct {
	Text           string  `json:"text"`
	FontFamily     string  `json:"fontFamily"`
	FontSize       int     `json:"fontSize"`
	Bold           bool    `json:"bold"`
	Italic         bool    `json:"italic"`
	Underline      bool    `json:"underline"`
	Color          string  `json:"color"`
	StrokeColor    string  `json:"strokeColor"`
	StrokeWidth    int     `json:"strokeWidth"`
	ShadowColor    string  `json:"shadowColor"`
	ShadowBlur     int     `json:"shadowBlur"`
	ShadowOffsetX  int     `json:"shadowOffsetX"`
	ShadowOffsetY  int     `json:"shadowOffsetY"`
	BgColor        string  `json:"bgColor"`
	BgPadding      int     `json:"bgPadding"`
	BgBorderRadius int     `json:"bgBorderRadius"`
	Align          string  `json:"align"`
	LetterSpacing  int     `json:"letterSpacing"`
	LineHeight     float64 `json:"lineHeight"`
	// Animation is a text-entry preset id: "" (none) | fade_in | fade_out |
	// typewriter | slide_left | slide_right | slide_top | slide_bottom |
	// bounce | pop | zoom_in | wipe. Applied over AnimationDuration seconds
	// at clip start; export falls back to fade_in for typewriter (preview
	// keeps the true effect). See internal/ffmpeg/filters/text.go.
	Animation         string  `json:"animation,omitempty"`
	AnimationDuration float64 `json:"animationDuration,omitempty"`
}

type StickerProps struct {
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Width    float64 `json:"width"`
	Height   float64 `json:"height"`
	Rotation float64 `json:"rotation"`
	Opacity  float64 `json:"opacity"`
	FlipH    bool    `json:"flipH"`
	FlipV    bool    `json:"flipV"`
}

type Keyframe struct {
	ID       string      `json:"id"`
	Time     float64     `json:"time"`
	Property string      `json:"property"`
	Value    interface{} `json:"value"`
	Easing   string      `json:"easing"`
}

type Transition struct {
	Type     string  `json:"type"`
	Duration float64 `json:"duration"`
}

type Clip struct {
	ID             string        `json:"id"`
	AssetID        string        `json:"assetId"`
	TrackID        string        `json:"trackId"`
	StartTime      float64       `json:"startTime"`
	Duration       float64       `json:"duration"`
	TrimStart      float64       `json:"trimStart"`
	TrimEnd        float64       `json:"trimEnd"`
	Speed          float64       `json:"speed"`
	Reversed       bool          `json:"reversed"`
	Volume         float64       `json:"volume"`
	Opacity        float64       `json:"opacity"`
	Transform      Transform     `json:"transform"`
	Color          ColorGrade    `json:"color"`
	Keyframes      []Keyframe    `json:"keyframes"`
	Transition     *Transition   `json:"transition,omitempty"`
	Normalize      bool          `json:"normalize"`
	NoiseReduction bool          `json:"noiseReduction"`
	TextProps      *TextProps    `json:"textProps,omitempty"`
	StickerProps   *StickerProps `json:"stickerProps,omitempty"`
}

type AssetType string

const (
	AssetVideo AssetType = "video"
	AssetAudio AssetType = "audio"
	AssetImage AssetType = "image"
	AssetGIF   AssetType = "gif"
)

type Asset struct {
	ID         string    `json:"id"`
	Path       string    `json:"path"`
	Type       AssetType `json:"type"`
	HasAudio   bool      `json:"hasAudio,omitempty"`
	Duration   float64   `json:"duration"`
	Width      int       `json:"width"`
	Height     int       `json:"height"`
	FPS        float64   `json:"fps"`
	Codec      string    `json:"codec"`
	Thumbnail  string    `json:"thumbnail"`
	Waveform   []float32 `json:"waveform"`
	FileSize   int64     `json:"fileSize"`
	ImportedAt time.Time `json:"importedAt"`
}

type RenderSettings struct {
	JobID        string  `json:"jobId"`
	OutputPath   string  `json:"outputPath"`
	Format       string  `json:"format"`
	Codec        string  `json:"codec"`
	Width        int     `json:"width"`
	Height       int     `json:"height"`
	FPS          float64 `json:"fps"`
	Bitrate      string  `json:"bitrate"`
	AudioBitrate string  `json:"audioBitrate"`
	CRF          int     `json:"crf"`
	Preset       string  `json:"preset"`
	StartTime    float64 `json:"startTime"`
	EndTime      float64 `json:"endTime"`
}

type RenderStatus string

const (
	RenderQueued    RenderStatus = "queued"
	RenderRendering RenderStatus = "rendering"
	RenderDone      RenderStatus = "done"
	RenderError     RenderStatus = "error"
	RenderCancelled RenderStatus = "cancelled"
)

type RenderProgress struct {
	JobID       string  `json:"jobId"`
	Percent     float64 `json:"percent"`
	CurrentTime float64 `json:"currentTime"`
	TotalTime   float64 `json:"totalTime"`
	FPS         float64 `json:"fps"`
	Status      string  `json:"status"`
	Error       string  `json:"error,omitempty"`
	OutputPath  string  `json:"outputPath,omitempty"`
}

type RecentProject struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updatedAt"`
	Thumbnail string    `json:"thumbnail,omitempty"`
}

type MediaInfo struct {
	Path       string  `json:"path"`
	Duration   float64 `json:"duration"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	FPS        float64 `json:"fps"`
	Codec      string  `json:"codec"`
	AudioCodec string  `json:"audioCodec"`
	FileSize   int64   `json:"fileSize"`
}

type FontInfo struct {
	Family string `json:"family"`
	Path   string `json:"path"`
	Style  string `json:"style"`
}

type FileFilter struct {
	Name       string   `json:"name"`
	Extensions []string `json:"extensions"`
}

type AssetThumbnailEvent struct {
	AssetID string `json:"assetId"`
	Data    string `json:"data"`
}

type AssetWaveformEvent struct {
	AssetID string    `json:"assetId"`
	Data    []float32 `json:"data"`
}
