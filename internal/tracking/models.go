package tracking

type TrackingMethod string

const (
	TrackStabilize TrackingMethod = "stabilize"
	TrackPoint     TrackingMethod = "point"
)

type TrackingSettings struct {
	Method    TrackingMethod `json:"method"`
	AssetID   string         `json:"assetId"`
	StartTime float64        `json:"startTime"`
	Duration  float64        `json:"duration"`
	// Point tracking specifics
	RegionX int `json:"regionX"`
	RegionY int `json:"regionY"`
	RegionW int `json:"regionW"`
	RegionH int `json:"regionH"`
	// Stabilization specifics
	Shaking  int `json:"shaking"`  // 0=low, 1=medium, 2=high
	Accuracy int `json:"accuracy"` // 0=fast, 1=accurate, 2=very accurate
}

type TrackingData struct {
	AssetID       string         `json:"assetId"`
	Method        TrackingMethod `json:"method"`
	Points        []TrackedPoint `json:"points"`
	TransformFile string         `json:"transformFile"`
	Confidence    float64        `json:"confidence"`
	FrameCount    int            `json:"frameCount"`
	Error         string         `json:"error,omitempty"`
}

type TrackedPoint struct {
	Frame int     `json:"frame"`
	Time  float64 `json:"time"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}
