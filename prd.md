# Gocut — Product Requirements Document

**Version:** 1.0  
**Status:** Draft  
**Last Updated:** June 2026  
**License:** Apache 2.0 (recommended)

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Problem Statement](#2-problem-statement)
3. [Goals & Non-Goals](#3-goals--non-goals)
4. [Target Users](#4-target-users)
5. [Tech Stack](#5-tech-stack)
6. [Architecture Overview](#6-architecture-overview)
7. [Data Models](#7-data-models)
8. [Feature Specifications](#8-feature-specifications)
   - 8.1 Project Management
   - 8.2 Media Import & Asset Pool
   - 8.3 Timeline Editor
   - 8.4 Canvas Preview Player
   - 8.5 Video Clip Operations
   - 8.6 Color Grading & Filters
   - 8.7 Text & Titles
   - 8.8 Audio Engine
   - 8.9 Transitions
   - 8.10 Stickers & Overlays
   - 8.11 Export & Render Engine
9. [Wails API Bindings](#9-wails-api-bindings)
10. [UI Layout Specification](#10-ui-layout-specification)
11. [MVP Scope (v0.1)](#11-mvp-scope-v01)
12. [Full Scope (v1.0)](#12-full-scope-v10)
13. [Performance Requirements](#13-performance-requirements)
14. [Technical Risks & Mitigations](#14-technical-risks--mitigations)
15. [Go Package Structure](#15-go-package-structure)
16. [Vue Component Tree](#16-vue-component-tree)
17. [Open Source Considerations](#17-open-source-considerations)
18. [Milestones & Timeline](#18-milestones--timeline)

---

## 1. Executive Summary

**Gocut ** is a cross-platform, open-source desktop video editor built with Go (Wails v2) and a Vue/JavaScript frontend. It targets content creators, short-form video makers, YouTubers, and developers who want a capable, privacy-respecting, fully offline video editing tool — a credible open-source desktop alternative to CapCut.

It is powered by FFmpeg for all media processing, uses SQLite for project persistence, and exposes a clean Wails RPC bridge between the Go backend and the Vue UI. There is no cloud dependency, no telemetry, and no account required.

---

## 2. Problem Statement

CapCut is the dominant tool for short-form video editing, but it is:

- Cloud-dependent and privacy-invasive (uploads footage to servers)
- Not open source
- Not available as a first-class desktop experience on Linux
- Subject to potential bans, regional restrictions, and policy changes
- Monetized through AI upsells behind paywalls

Content creators and developers need a **powerful, local, open-source** alternative that respects their data, runs on any OS, and can be extended or self-hosted.

---

## 3. Goals & Non-Goals

### Goals

- Feature parity with CapCut's core desktop editing workflow
- 100% offline — no network calls required to use
- Cross-platform: Windows 10+, macOS 12+, Ubuntu 20.04+
- Fast UI with non-blocking Go backend processing
- Clean, well-documented codebase suitable for open source contributions
- Permissive license (Apache 2.0) to allow commercial forks

### Non-Goals

- Mobile app (iOS/Android)
- Cloud sync or collaboration features
- AI-generated content (text-to-video, AI effects) in v1.0
- Professional NLE features (multi-cam, color nodes, Resolve-level grading)
- Plugin marketplace in v1.0

---

## 4. Target Users

| Persona | Description | Key Need |
|---|---|---|
| Short-form Creator | Makes YouTube Shorts, TikTok-style content | Fast cuts, text overlays, music sync |
| Indie YouTuber | Produces 5–30 min videos solo | Stable timeline, export quality control |
| Developer / Privacy User | Doesn't trust cloud tools | Fully local, open source |
| Linux User | Underserved by CapCut | Native Linux support |
| Educator / Lecturer | Records and edits screen + webcam | Simple trimming, subtitle overlays |

---

## 5. Tech Stack

### Backend (Go)

| Component | Library / Tool |
|---|---|
| Desktop Shell | Wails v2 |
| Media Processing | FFmpeg via `os/exec` + structured argument builder |
project db 
github.com/ncruces/go-sqlite3
| Audio Analysis | `github.com/go-audio/audio` + custom MFCC pipeline |
| File Watching | `github.com/fsnotify/fsnotify` |
| Thumbnail Cache | Disk-based LRU cache (custom, pure Go) |
| Background Jobs | Goroutines + buffered channels + context cancellation |
| Image Processing | `github.com/disintegration/imaging` |

### Frontend (Vue + JavaScript)

| Component | Library |
|---|---|
| Build Tool | Vite |
| Framework | Vue 3 + JavaScript |
| State Management | Pinia |
| UI Components | Tailwind CSS + shadcn-vue |
| Animations | GSAP (via @vueuse/motion or Vue transitions) |
| Waveform Display | WaveSurfer.js |
| Canvas Editing | Konva.js (text, stickers, overlays) |
| Timeline | Custom Vue components (no third-party dependency) |
| Icons | Lucide (Vue) |

---

## 6. Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Wails Desktop Shell                   │
│                                                         │
│  ┌──────────────────┐    ┌───────────────────────────┐  │
│  │   Vue Frontend   │◄──►│     Go Backend (RPC)      │  │
│  │                  │    │                           │  │
│  │  - Timeline UI   │    │  - FFmpeg Executor        │  │
│  │  - Canvas Preview│    │  - Project Manager        │  │
│  │  - Asset Pool    │    │  - Render Queue           │  │
│  │  - Properties    │    │  - Thumbnail Cache        │  │
│  │  - Export Dialog │    │  - Waveform Analyzer      │  │
│  └──────────────────┘    │  - File Watcher           │  │
│                          └───────────┬───────────────┘  │
│                                      │                  │
│                          ┌───────────▼───────────────┐  │
│                          │    SQLite (Project DB)    │  │
│                          │    Disk Cache (Thumbs)    │  │
│                          │    FFmpeg Process Pool    │  │
│                          └───────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

### Communication Patterns

- **Frontend → Backend:** Wails Go bindings (synchronous RPC calls)
- **Backend → Frontend:** `runtime.EventsEmit` for async push (render progress, autosave, watcher events)
- **Heavy operations:** Always run in Go goroutines; frontend shows loading state immediately and receives result via event

---

## 7. Data Models

### ProjectJSON

```go
type Project struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
    Duration    float64   `json:"duration"` // seconds
    AspectRatio string    `json:"aspectRatio"` // "16:9", "9:16", "1:1", "4:3"
    Resolution  Resolution `json:"resolution"`
    FPS         float64   `json:"fps"`
    Timeline    Timeline  `json:"timeline"`
    Assets      []Asset   `json:"assets"`
    Settings    ProjectSettings `json:"settings"`
}

type Resolution struct {
    Width  int `json:"width"`
    Height int `json:"height"`
}

type ProjectSettings struct {
    BackgroundColor string `json:"backgroundColor"`
    AutoSave        bool   `json:"autoSave"`
    AutoSaveInterval int   `json:"autoSaveIntervalSeconds"`
}
```

### Timeline & Tracks

```go
type Timeline struct {
    Tracks   []Track  `json:"tracks"`
    Duration float64  `json:"duration"`
}

type Track struct {
    ID    string    `json:"id"`
    Type  TrackType `json:"type"` // "video" | "audio" | "text" | "sticker" | "fx"
    Clips []Clip    `json:"clips"`
    Muted bool      `json:"muted"`
    Locked bool     `json:"locked"`
    Volume float64  `json:"volume"` // audio tracks only
}

type TrackType string
const (
    TrackVideo   TrackType = "video"
    TrackAudio   TrackType = "audio"
    TrackText    TrackType = "text"
    TrackSticker TrackType = "sticker"
    TrackFX      TrackType = "fx"
)
```

### Clip

```go
type Clip struct {
    ID         string    `json:"id"`
    AssetID    string    `json:"assetId"`
    TrackID    string    `json:"trackId"`
    StartTime  float64   `json:"startTime"`  // position on timeline (seconds)
    Duration   float64   `json:"duration"`   // how long it plays
    TrimStart  float64   `json:"trimStart"`  // offset into source asset
    TrimEnd    float64   `json:"trimEnd"`    // end offset in source asset
    Speed      float64   `json:"speed"`      // 1.0 = normal
    Reversed   bool      `json:"reversed"`
    Volume     float64   `json:"volume"`     // 0.0 - 2.0
    Opacity    float64   `json:"opacity"`    // 0.0 - 1.0
    Transform  Transform `json:"transform"`
    Color      ColorGrade `json:"color"`
    Keyframes  []Keyframe `json:"keyframes"`
    Transition *Transition `json:"transition,omitempty"` // at clip boundary
    TextProps  *TextProps  `json:"textProps,omitempty"`
    StickerProps *StickerProps `json:"stickerProps,omitempty"`
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
```

### Asset

```go
type Asset struct {
    ID        string    `json:"id"`
    Path      string    `json:"path"`      // absolute or relative to project
    Type      AssetType `json:"type"`      // "video" | "audio" | "image" | "gif"
    Duration  float64   `json:"duration"`  // 0 for images
    Width     int       `json:"width"`
    Height    int       `json:"height"`
    FPS       float64   `json:"fps"`
    Codec     string    `json:"codec"`
    Thumbnail string    `json:"thumbnail"` // base64 or cache key
    Waveform  []float32 `json:"waveform"`  // normalized amplitude samples
    FileSize  int64     `json:"fileSize"`
    ImportedAt time.Time `json:"importedAt"`
}
```

### Keyframe

```go
type Keyframe struct {
    ID        string      `json:"id"`
    Time      float64     `json:"time"`     // seconds from clip start
    Property  string      `json:"property"` // "x", "y", "scaleX", "opacity", etc.
    Value     interface{} `json:"value"`
    Easing    string      `json:"easing"`   // "linear" | "ease-in" | "ease-out" | "ease-in-out"
}
```

### RenderSettings

```go
type RenderSettings struct {
    JobID      string `json:"jobId"`
    OutputPath string `json:"outputPath"`
    Format     string `json:"format"`     // "mp4" | "webm" | "gif" | "mp3"
    Codec      string `json:"codec"`      // "h264" | "h265" | "vp9" | "av1"
    Width      int    `json:"width"`
    Height     int    `json:"height"`
    FPS        float64 `json:"fps"`
    Bitrate    string `json:"bitrate"`    // "auto" or e.g. "8000k"
    AudioBitrate string `json:"audioBitrate"`
    CRF        int    `json:"crf"`        // 0-51 for H.264
    Preset     string `json:"preset"`     // "ultrafast" to "veryslow"
    StartTime  float64 `json:"startTime"` // partial export
    EndTime    float64 `json:"endTime"`
}

type RenderProgress struct {
    JobID      string  `json:"jobId"`
    Percent    float64 `json:"percent"`
    CurrentTime float64 `json:"currentTime"`
    TotalTime  float64 `json:"totalTime"`
    FPS        float64 `json:"fps"`
    Status     string  `json:"status"` // "queued" | "rendering" | "done" | "error" | "cancelled"
    Error      string  `json:"error,omitempty"`
    OutputPath string  `json:"outputPath,omitempty"`
}
```

---

## 8. Feature Specifications

### 8.1 Project Management

**User Stories:**
- As a creator, I can create a new project with a name, aspect ratio, and resolution
- As a creator, I can save and load projects from disk as portable `.Gocut` files (JSON)
- As a creator, I have auto-save so I never lose work
- As a creator, I see recent projects on the home screen

**Acceptance Criteria:**
- New project dialog lets user pick: name, aspect ratio (16:9, 9:16, 1:1, 4:3, custom), resolution (4K/1080p/720p/480p), FPS (24/25/30/60)
- Projects saved as a single JSON file; assets referenced by relative paths
- Auto-save triggers every 60 seconds (configurable) and on app focus-out
- Recent projects list shows last 10, with thumbnail preview (first frame of first video clip)
- Project can be renamed from the title bar (double-click)
- Unsaved changes prompt on close

---

### 8.2 Media Import & Asset Pool

**User Stories:**
- As a creator, I can import videos, audio, and images by file picker or drag-and-drop
- As a creator, I can browse imported assets in a panel with thumbnail previews
- As a creator, I can remove assets from the pool (with warning if in use)

**Acceptance Criteria:**
- Supported import formats: MP4, MOV, AVI, MKV, WebM (video); MP3, WAV, AAC, FLAC (audio); PNG, JPG, JPEG, GIF, WebP (image)
- Drag-and-drop from OS file manager into the asset pool
- Thumbnail generated asynchronously via FFmpeg `ffmpeg -ss 0 -i input -frames:v 1 thumb.jpg`
- Waveform data generated asynchronously for audio/video with audio tracks
- Asset pool shows: thumbnail, name, duration badge, file type badge
- Assets can be dragged from pool directly onto the timeline
- Asset pool supports search/filter by type
- Unsupported file shows a clear error toast (not silent fail)
- Duplicate detection by file path (warn, don't block)

---

### 8.3 Timeline Editor

**User Stories:**
- As a creator, I can place clips on a multi-track timeline and arrange them visually
- As a creator, I can trim, split, and rearrange clips with mouse interactions
- As a creator, the timeline shows me a visual scrubber synced to the preview player

**Acceptance Criteria:**

**Tracks:**
- Support minimum 1 video track, 2 audio tracks, 1 text track, 1 sticker track, 1 FX track in MVP
- Tracks can be added/removed dynamically in v1.0
- Each track has: mute toggle, lock toggle, label, type icon

**Clip Operations:**
- Drag clip horizontally to reposition (shows snap guides)
- Drag left/right edge of clip to trim (TrimStart/TrimEnd)
- Right-click context menu: Split at Playhead, Duplicate, Delete, Detach Audio, Speed
- Double-click clip to open properties in the right panel
- Ripple delete: when enabled, subsequent clips shift left after delete
- Select multiple clips with Shift+Click or drag-select rectangle
- Copy/paste clips (Ctrl+C / Ctrl+V)

**Playhead:**
- Clicking on the ruler seeks to that position
- Dragging playhead scrubs (throttled to 150ms, triggers frame extraction)
- Playhead shows current time as `00:00:00.000`
- Keyboard: Space = play/pause, Left/Right arrows = ±1 frame, J/K/L = shuttle

**Zoom:**
- Mouse wheel on timeline zooms in/out (min: 1px/s, max: 200px/s)
- Zoom level stored in UI state (not project)
- Zoom-to-fit button in timeline toolbar

**Snap:**
- Snap to clip edges (on by default, toggle with S)
- Snap to playhead
- Snap to second marks
- Visual yellow snap indicator line

**Thumbnails:**
- Video clips show frame thumbnails at sufficient zoom levels
- Audio clips show waveform visualization
- Text clips show text label preview

---

### 8.4 Canvas Preview Player

**User Stories:**
- As a creator, I can see a real-time preview of my composition as I edit
- As a creator, I can scrub through the timeline and see accurate frames

**Acceptance Criteria:**

**Playback:**
- Play/pause with Space bar
- Real-time playback via frame extraction: Go extracts frames ahead of playhead into a ring buffer (5 frames)
- Supported preview resolutions: Full (project res), 1/2, 1/4
- Frame-accurate seeking via FFmpeg: `ffmpeg -ss {time} -i input -frames:v 1 -f image2pipe -`
- Audio playback synced to video (via Web Audio API in frontend)
- Loop toggle button
- Volume control for master preview

**Canvas:**
- Konva.js canvas layer on top for interactive text/sticker manipulation
- Click on text/sticker overlay to select (shows resize/rotate handles)
- Canvas respects aspect ratio; letterbox/pillarbox shown for mismatched content

**Controls:**
- Play, Pause, Step Back (1 frame), Step Forward (1 frame)
- Timeline mini-scrubber under the preview
- Current time / total duration display
- Fullscreen preview mode

---

### 8.5 Video Clip Operations

**User Stories:**
- As a creator, I can apply basic transformations to any video clip
- As a creator, I can change clip speed, reverse it, or crop it

**Acceptance Criteria:**

| Feature | Detail |
|---|---|
| Speed | 0.1x – 16x, shown as badge on clip. Pitch correction toggle (via `atempo` filter). |
| Reverse | Triggers `ffmpeg -vf reverse -af areverse` on clip segment, cached to temp file |
| Crop | Free crop with handles on canvas, or aspect ratio presets |
| Rotate | Free rotation input (degrees) or snap: 90°, 180°, 270° |
| Flip | Horizontal / Vertical toggles |
| Opacity | 0–100% slider, applied via FFmpeg `format=rgba,colorchannelmixer=aa=VALUE` |
| Scale & Position | X/Y/W/H inputs in properties panel, or drag on canvas |
| Keyframes | Any transform property can have keyframes; show keyframe diamond on clip bar |

FFmpeg filter for transform:
```
-vf "scale=W:H,pad=PW:PH:X:Y,rotate=R*PI/180,hflip,vflip,format=yuva420p,colorchannelmixer=aa=OPACITY"
```

---

### 8.6 Color Grading & Filters

**User Stories:**
- As a creator, I can adjust the look and feel of video clips with color tools
- As a creator, I can apply one-click filter presets like CapCut's filter strip

**Acceptance Criteria:**

**Manual Adjustments (per-clip):**

| Control | Range | FFmpeg Filter |
|---|---|---|
| Brightness | -100 to +100 | `eq=brightness=V` |
| Contrast | -100 to +100 | `eq=contrast=V` |
| Saturation | -100 to +100 | `eq=saturation=V` |
| Hue | -180 to +180 | `hue=h=V` |
| Sharpness | 0 to 100 | `unsharp=5:5:V` |
| Vignette | 0 to 100 | `vignette=angle=PI/4*V` |
| Noise / Grain | 0 to 100 | `noise=alls=V` |
| Blur | 0 to 20px | `boxblur=V` |

**Filter Presets:**
- Horizontal scrollable strip of preset thumbnails (CapCut-style)
- Built-in presets (min 12): Natural, Cinema, Warm, Cool, Vintage, B&W, Fade, Vivid, Matte, Golden Hour, Cyberpunk, Soft
- Each preset is a named set of the manual adjustment values above
- Applying a preset fills the sliders; user can still adjust after

**Chroma Key (Basic):**
- Color picker to select key color (default green)
- Similarity and blend sliders
- FFmpeg: `chromakey=color=HEX:similarity=S:blend=B`

---

### 8.7 Text & Titles

**User Stories:**
- As a creator, I can add animated text overlays to my video
- As a creator, I can customize font, size, color, and style of text

**Acceptance Criteria:**

**Adding Text:**
- "Add Text" button creates a text clip on the text track at playhead position
- Default duration: 3 seconds (draggable to adjust)
- Text is editable by double-clicking on the canvas or in the properties panel

**Text Styling:**
- Font family: system fonts + bundled open-licensed fonts (min 20: Roboto, Montserrat, Playfair Display, etc.)
- Font size: 8–300px
- Bold, Italic, Underline toggles
- Text color (color picker)
- Stroke color + stroke width
- Shadow: color, blur, offsetX, offsetY
- Background box: color, padding, border-radius
- Text alignment: left, center, right
- Letter spacing, line height

**Text Animation Presets (min 12):**
- None, Fade In, Fade Out, Typewriter, Slide In Left/Right/Top/Bottom, Bounce, Pop, Zoom In, Wipe
- Animation duration: configurable (default 0.5s)
- Preview animation in properties panel with mini loop

**Emoji Support:**
- Emoji rendered via canvas font fallback (system emoji font)
- Emoji picker button in text toolbar

**Rendering:**
- Text rendered via FFmpeg `drawtext` filter or composited as PNG overlay via Konva.js export
- For complex styled text: export frame as PNG from Konva, overlay with FFmpeg `overlay` filter

---

### 8.8 Audio Engine

**User Stories:**
- As a creator, I can control the volume of each clip independently
- As a creator, I can add background music and see waveforms in the timeline

**Acceptance Criteria:**

**Per-Clip Audio:**
- Volume slider 0–200% per clip
- Fade in / fade out handles (drag the edge of the audio region within the clip)
- FFmpeg: `afade=t=in:st=0:d=0.5,afade=t=out:st=END-0.5:d=0.5`
- Mute toggle per clip

**Volume Keyframes:**
- Keyframe button in audio properties to add volume keyframes
- Volume envelope drawn over waveform in timeline
- FFmpeg: `volume=enable='between(t,0,1)':volume=0.5`

**Background Music Track:**
- Dedicated audio track for BGM
- Loop toggle: loops the audio clip to fill the project duration
- BGM duck: optional sidechain duck (reduce BGM volume when main audio is present)

**Waveform Visualization:**
- Waveform extracted via `ffmpeg -filter_complex aformat=channel_layouts=mono,compand -ac 1 -f f32le`
- Downsampled to ~500 samples for rendering
- Drawn as SVG path in timeline clip background

**Noise Reduction:**
- Toggle button: applies `afftdn=nf=-25` FFmpeg filter
- Processed as a non-destructive flag on the clip (rendered only on export)

**Audio Normalization:**
- Per-track loudness normalization toggle: `loudnorm=I=-16:TP=-1.5:LRA=11`

---

### 8.9 Transitions

**User Stories:**
- As a creator, I can add transitions between clips to smooth the edit
- As a creator, I can preview transitions before applying them

**Acceptance Criteria:**

**Built-in Transitions (min 12):**

| Name | FFmpeg Implementation |
|---|---|
| Cut | No filter (hard cut, default) |
| Fade | `xfade=transition=fade` |
| Dissolve | `xfade=transition=dissolve` |
| Wipe Left | `xfade=transition=wipeleft` |
| Wipe Right | `xfade=transition=wiperight` |
| Slide Left | `xfade=transition=slideleft` |
| Slide Right | `xfade=transition=slideright` |
| Zoom In | `xfade=transition=zoomin` |
| Flip H | `xfade=transition=hflip` |
| Blur | Custom: blend between blurred versions |
| Circle Open | `xfade=transition=circleopen` |
| Pixelize | `xfade=transition=pixelize` |

**Transition Properties:**
- Duration: 0.1s – 2.0s (default 0.5s)
- Transition stored on the leading clip's `transition` field
- Shown as an overlapping notch on the timeline between two clips

**UI:**
- Transitions panel (left panel tab) shows scrollable grid of transition thumbnails
- Drag transition onto clip boundary to apply
- Double-click applied transition to edit duration
- Preview: clicking a transition thumbnail shows a 3-second looped animation preview

---

### 8.10 Stickers & Overlays

**User Stories:**
- As a creator, I can add stickers and image overlays on top of my video
- As a creator, I can animate sticker position and scale

**Acceptance Criteria:**

- Sticker clips live on a dedicated sticker track
- Import custom PNG or GIF as sticker via file picker
- Built-in sticker pack: min 30 open-licensed SVG/PNG stickers (emoji-style, shapes, arrows, borders)
- Sticker properties: X, Y, Width, Height, Rotation, Opacity, Flip H/V
- Sticker supports keyframes on position, scale, opacity, rotation
- GIF stickers loop for the duration of the clip
- Rendered via FFmpeg `overlay` or `movie` source filter

---

### 8.11 Export & Render Engine

**User Stories:**
- As a creator, I can export my project to a video file in various formats and resolutions
- As a creator, the export runs in the background while I continue editing

**Acceptance Criteria:**

**Export Formats:**
- MP4 (H.264 / H.265)
- WebM (VP9)
- GIF (with palette optimization)
- MP3 / AAC (audio-only export)

**Resolution Presets:**
- 4K (3840×2160), 1080p (1920×1080), 720p (1280×720), 480p (854×480)
- Aspect ratio variants: 16:9, 9:16, 1:1, 4:3
- Custom resolution input

**Quality Control:**
- CRF slider (0-51 for H.264, lower = better quality)
- Preset: ultrafast → veryslow (speed vs. size tradeoff)
- Manual bitrate input (or "Auto")
- Audio bitrate: 64k, 128k, 192k, 320k

**Background Rendering:**
- Render starts a Go goroutine; UI is not blocked
- Progress bar shows: percent, current time / total time, current FPS
- Cancel button stops render mid-way (kills FFmpeg process via `cmd.Process.Kill()`)
- Render queue supports multiple queued jobs (v1.0)
- Completion notification via Wails event → OS toast (via Wails `runtime.WindowExecJS`)

**Partial Export:**
- User can set in/out points on timeline for range export
- Useful for exporting a specific section

**FFmpeg Filter Graph Construction (Go):**
The Go backend assembles `-filter_complex` strings from the project Timeline:

```
[0:v]trim=start=TRIM_START:end=TRIM_END,setpts=PTS-STARTPTS,
     scale=W:H,rotate=R,
     eq=brightness=B:contrast=C:saturation=S,
     lut3d=lut.cube
[v0];

[1:v]trim=...[v1];

[v0][v1]xfade=transition=fade:duration=0.5:offset=4.5[vout];

[0:a]afade=t=in:st=0:d=0.5,volume=0.8[a0];
[1:a]volume=1.0[a1];
[a0][a1]amix=inputs=2[aout]
```

---

## 9. Wails API Bindings

All methods are on the `App` struct and exposed to the frontend via Wails `Bind`.

### Project

```go
func (a *App) NewProject(settings ProjectSettings) (*Project, error)
func (a *App) SaveProject(project Project) error
func (a *App) LoadProject(path string) (*Project, error)
func (a *App) GetRecentProjects() ([]RecentProject, error)
func (a *App) ExportProjectFile(project Project) (string, error) // opens save dialog
```

### Media

```go
func (a *App) ImportMedia(paths []string) ([]Asset, error)
func (a *App) ExtractThumbnail(assetID string, timeMs int) (string, error)    // returns base64 PNG
func (a *App) ExtractWaveform(assetID string) ([]float32, error)
func (a *App) GetMediaInfo(path string) (*MediaInfo, error)                   // codec, dimensions, fps, duration
func (a *App) GenerateThumbnailStrip(assetID string, count int) ([]string, error) // for timeline thumbnails
```

### Render

```go
func (a *App) StartRender(project Project, settings RenderSettings) (string, error) // returns jobID
func (a *App) GetRenderProgress(jobID string) (*RenderProgress, error)
func (a *App) CancelRender(jobID string) error
func (a *App) GetRenderQueue() ([]RenderProgress, error)
func (a *App) OpenOutputFolder(path string) error
```

### Preview

```go
func (a *App) GetPreviewFrame(project Project, timeSeconds float64, width int, height int) (string, error) // base64 JPEG
func (a *App) PreloadFrames(project Project, startTime float64, count int) error // pre-warms frame cache
func (a *App) ClearPreviewCache() error
```

### Utility

```go
func (a *App) CheckFFmpegInstalled() (string, error) // returns version string or error
func (a *App) OpenFilePicker(filters []FileFilter) ([]string, error)
func (a *App) GetAppVersion() string
func (a *App) GetSystemFonts() ([]FontInfo, error)
```

### Wails Events (Go → Frontend)

```go
runtime.EventsEmit(ctx, "render:progress", RenderProgress{...})
runtime.EventsEmit(ctx, "render:complete", RenderProgress{...})
runtime.EventsEmit(ctx, "render:error", RenderProgress{...})
runtime.EventsEmit(ctx, "project:autosaved", time.Now().Unix())
runtime.EventsEmit(ctx, "asset:thumbnailReady", AssetThumbnailEvent{...})
runtime.EventsEmit(ctx, "asset:waveformReady", AssetWaveformEvent{...})
runtime.EventsEmit(ctx, "ffmpeg:notFound", nil)
```

---

## 10. UI Layout Specification

```
┌────────────────────────────────────────────────────────────────┐
│  TOP BAR: [Logo] [Project Name▼] [Undo][Redo]  [Export ▶] [⚙] │
├──────────────────────────────┬─────────────────────────────────┤
│  LEFT PANEL (240px)          │  CENTER: Canvas Preview         │
│  ┌──────────────────────┐    │  ┌─────────────────────────┐    │
│  │ [Media][Audio][Text] │    │  │                         │    │
│  │ [FX][Stickers][Trans]│    │  │    16:9 Canvas           │    │
│  ├──────────────────────┤    │  │    (Konva + Preview)     │    │
│  │                      │    │  │                         │    │
│  │  Asset / Panel Grid  │    │  └─────────────────────────┘    │
│  │  (thumbnails/items)  │    │  [◀◀] [◀] [▶/⏸] [▶] [▶▶]      │
│  │                      │    │  00:00:12.450 / 00:01:30.000    │
│  └──────────────────────┘    │                                 │
├──────────────────────────────┴─────────────────────────────────┤
│  TIMELINE (bottom, resizable height ~200px)                    │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ [+Track] [Zoom─────+] [Snap●] [Razor] [Ripple]         │   │
│  ├────────┬────────────────────────────────────────────────┤   │
│  │ 🎬 V1  │ ████CLIP A████████████ ████CLIP B████ ...     │   │
│  │ 🔊 A1  │ ~~~waveform~~~~~~~~~~~ ~~~waveform~~~ ...     │   │
│  │ 🔊 A2  │                        ~~~BGM~~~~~~~~~...     │   │
│  │ 💬 T1  │         [TEXT CLIP]                           │   │
│  │ ⭐ S1  │                  [STICKER]                    │   │
│  └────────┴────────────────────────────────────────────────┘   │
├────────────────────────────────────────────────────────────────┤
│  RIGHT PANEL (280px) — Properties Inspector                    │
│  (shows selected clip's transform, color, audio, text props)   │
└────────────────────────────────────────────────────────────────┘
```

**Design System:**
- Dark theme: `#0F0F0F` background, `#1A1A1A` panel bg, `#2A2A2A` border
- Accent: `#00D4FF` (cyan-blue, CapCut-inspired)
- Text: `#E8E8E8` primary, `#888888` secondary
- Font: `DM Sans` (UI), `JetBrains Mono` (timecode)
- Border radius: 6px components, 3px inputs

---

## 11. MVP Scope (v0.1)

The following features constitute a shippable v0.1:

**Included:**
- Project create/save/load
- Media import (video + audio + image), thumbnail generation
- Single video track + single audio track on timeline
- Clip placement, trim, split, move, delete
- Basic preview player (frame extraction, 15fps limit)
- Trim, speed (0.5x / 1x / 2x presets only), opacity
- Basic color: brightness, contrast, saturation sliders
- Text overlay (single font, no animation)
- Export: MP4 H.264, 1080p, 720p only
- Background render with progress bar and cancel
- FFmpeg bundled (Windows/macOS) or PATH-detected (Linux)
- Auto-save

**Not in MVP:**
- Keyframe animation
- Transitions
- LUT / filter presets
- Audio keyframes / fade handles
- Stickers
- GIF export
- Chroma key
- Waveform visualization
- Multiple video tracks

---

## 12. Full Scope (v1.0)

Everything in MVP, plus:

- All timeline features (snap, ripple, multi-select, razor tool)
- Full keyframe animation system
- All 12 transitions with xfade
- Complete color grading suite (all sliders + LUT + presets)
- Text animations (12 presets)
- Sticker track with GIF support
- Audio waveform in timeline
- Audio fade handles and keyframes
- Volume envelope editing
- Noise reduction toggle
- Chroma key
- Export: all formats (WebM, GIF, MP3), all resolutions
- Render queue (multiple jobs)
- Multiple video + audio tracks
- System font loading
- Waveform extraction
- Filter preset strip (12 presets)

---

## 13. Performance Requirements

| Metric | Target |
|---|---|
| Timeline drag response | < 16ms (60fps DOM) |
| Playhead scrub → frame shown | < 150ms |
| Thumbnail generation (per clip) | < 2s (async) |
| Waveform generation | < 3s (async, background) |
| Project save | < 500ms |
| Project load (incl. assets) | < 2s |
| Export speed | ≥ 80% of native FFmpeg speed |
| App cold start | < 3s |
| Memory usage (idle) | < 200MB |
| Memory usage (during export) | < 600MB |

---

## 14. Technical Risks & Mitigations

| Risk | Severity | Mitigation |
|---|---|---|
| Real-time preview performance | High | Frame extraction caching with LRU eviction; preview at 1/2 or 1/4 resolution; pre-fetch 3 frames ahead of playhead |
| FFmpeg filter graph complexity | High | Dedicated `internal/ffmpeg/builder.go` that assembles filter_complex incrementally and is unit-tested for each clip type independently |
| FFmpeg not installed on Linux | Medium | Check at startup; show modal with install instructions; bundle static FFmpeg binary as optional download |
| Large project JSON (many clips) | Medium | Keep only IDs in Timeline, lazy-load clip full properties on selection |
| Audio sync drift during preview | Medium | Use Web Audio API clock as source of truth; resync every 500ms |
| LUT file handling (binary .cube) | Low | Parse .cube in Go, pass parsed table to FFmpeg via pipe or temp file |
| Wails WebView rendering on Linux | Medium | Test on both WebKit2GTK and Chromium; document requirements clearly |
| GIF export quality | Low | Use FFmpeg 2-pass palette generation (`palettegen` + `paletteuse`) |
| Cross-platform font discovery | Low | Use `golang.org/x/font` paths + OS-specific font dirs; fallback to bundled fonts |
| CGo dependency avoidance | Low | Use `database/sql` (stdlib) with a pure-Go SQLite driver like `modernc.org/sqlite` (avoid CGo). If using `github.com/ncruces/go-sqlite3`, prefer configurations that avoid/limit CGo; otherwise isolate it to optional build tags so non-CGo builds still work. |

---

## 15. Go Package Structure

```
Gocut/
├── cmd/
│   └── Gocut/
│       └── main.go              # Wails entry point
├── internal/
│   ├── app/
│   │   └── app.go               # App struct, Wails bindings
│   ├── ffmpeg/
│   │   ├── executor.go          # FFmpeg process management
│   │   ├── builder.go           # filter_complex string builder
│   │   ├── probe.go             # ffprobe wrapper (media info)
│   │   ├── thumbnail.go         # frame extraction
│   │   ├── waveform.go          # audio waveform extraction
│   │   └── filters/
│   │       ├── color.go         # color grade filter strings
│   │       ├── transform.go     # scale/rotate/crop filters
│   │       ├── audio.go         # audio filter strings
│   │       ├── text.go          # drawtext filter
│   │       └── transition.go    # xfade filter
│   ├── project/
│   │   ├── model.go             # Project, Timeline, Clip, Asset data models
│   │   ├── store.go             # SQLite project storage
│   │   ├── manager.go           # project CRUD operations
│   │   └── autosave.go          # fsnotify-based autosave
│   ├── render/
│   │   ├── queue.go             # render job queue (goroutines + channels)
│   │   ├── job.go               # individual render job
│   │   └── compositor.go        # assembles full FFmpeg command from project
│   ├── cache/
│   │   ├── thumbnail.go         # disk LRU cache for thumbnails
│   │   └── frame.go             # in-memory ring buffer for preview frames
│   ├── media/
│   │   ├── importer.go          # file import, validation, asset creation
│   │   └── info.go              # media metadata parsing
│   └── fonts/
│       └── scanner.go           # OS font directory scanning
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── timeline/
│   │   │   │   ├── Timeline.vue
│   │   │   │   ├── Track.vue
│   │   │   │   ├── Clip.vue
│   │   │   │   ├── Playhead.vue
│   │   │   │   └── Ruler.vue
│   │   │   ├── preview/
│   │   │   │   ├── PreviewPlayer.vue
│   │   │   │   └── Canvas.vue
│   │   │   ├── panels/
│   │   │   │   ├── AssetPool.vue
│   │   │   │   ├── TransitionsPanel.vue
│   │   │   │   ├── TextPanel.vue
│   │   │   │   ├── StickerPanel.vue
│   │   │   │   └── AudioPanel.vue
│   │   │   ├── inspector/
│   │   │   │   ├── Inspector.vue
│   │   │   │   ├── TransformProps.vue
│   │   │   │   ├── ColorProps.vue
│   │   │   │   ├── AudioProps.vue
│   │   │   │   └── TextProps.vue
│   │   │   ├── export/
│   │   │   │   ├── ExportDialog.vue
│   │   │   │   └── RenderProgress.vue
│   │   │   └── common/
│   │   │       ├── Slider.vue
│   │   │       ├── ColorPicker.vue
│   │   │       ├── Toast.vue
│   │   │       └── Modal.vue
│   │   ├── stores/
│   │   │   ├── projectStore.js
│   │   │   ├── timelineStore.js
│   │   │   ├── playerStore.js
│   │   │   └── uiStore.js
│   │   ├── composables/
│   │   │   ├── useWailsEvents.js
│   │   │   ├── usePlayer.js
│   │   │   └── useTimeline.js
│   │   ├── lib/
│   │   │   ├── wails.js        # wails bindings wrapper
│   │   │   └── time.js         # timecode formatting utilities
│   │   └── App.vue
│   ├── package.json
│   └── vite.config.js
├── assets/
│   ├── stickers/                # bundled open-license sticker PNGs
│   ├── fonts/                   # bundled open-license fonts
│   └── luts/                    # bundled sample LUT files
├── wails.json
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## 16. Vue Component Tree

```
App
├── HomeScreen (no project open)
│   ├── RecentProjectsList
│   └── NewProjectDialog
└── EditorLayout (project open)
    ├── TopBar
    │   ├── ProjectNameInput
    │   ├── UndoRedoButtons
    │   └── ExportButton → ExportDialog
    ├── LeftPanel
    │   ├── PanelTabs [Media, Audio, Text, Stickers, FX, Transitions]
    │   ├── AssetPool
    │   │   ├── AssetGrid
    │   │   └── AssetCard (thumbnail + metadata)
    │   ├── TransitionsPanel
    │   │   └── TransitionCard × N
    │   ├── TextPanel (templates / styles)
    │   └── StickerPanel
    │       └── StickerGrid
    ├── CenterPanel
    │   ├── PreviewPlayer
    │   │   ├── PreviewCanvas (Konva.js overlay)
    │   │   ├── FrameImage (extracted frame)
    │   │   └── PlayerControls
    │   └── TimelinePanel
    │       ├── TimelineToolbar
    │       ├── TimelineRuler
    │       ├── TrackList
    │       │   └── Track × N
    │       │       └── Clip × N
    │       └── Playhead
    └── RightPanel (Inspector)
        ├── ClipInspector (when clip selected)
        │   ├── TransformSection
        │   ├── ColorSection
        │   ├── AudioSection
        │   └── TextSection (text clips only)
        └── EmptyState (no selection)
```

### Pinia Stores

```js
// projectStore: current project JSON, dirty flag, save/load actions
// timelineStore: selectedClips, zoom level, scrollX, snap enabled
// playerStore: currentTime, isPlaying, volume, previewQuality
// uiStore: activePanel, activeInspectorTab, renderJobs, toasts
```

---

## 17. Open Source Considerations

**License:** Apache 2.0
- Allows commercial use and forks
- Compatible with FFmpeg (LGPL) when dynamically linked
- Compatible with all chosen Go and npm dependencies

**FFmpeg Licensing:**
- Distribute as a wrapper that calls system FFmpeg (PATH detection)
- Optionally offer pre-built static FFmpeg binary as a release artifact
- Do NOT statically compile FFmpeg into the binary (LGPL compliance)
- Include FFmpeg attribution in README and About dialog

**README structure:**
```
# Gocut 
> Open-source, offline video editor for desktop. Built with Go + Wails.

## Features | Screenshots | Download | Build from Source
## FFmpeg Installation
## Contributing
## License
```

**Contribution Guidelines:**
- CONTRIBUTING.md with: PR process, issue templates, coding style (gofmt, golangci-lint)
- GitHub Actions CI: `go build`, `go test ./...`, frontend lint on PRs
- Issue labels: `good first issue`, `bug`, `feature`, `performance`, `platform/linux`

---

## 18. Milestones & Timeline

| Milestone | Scope | Est. Duration |
|---|---|---|
| M0: Scaffold | Wails project, Go/Vue skeleton, FFmpeg detection, project create/save/load | 2 weeks |
| M1: Import & Asset Pool | Media import, thumbnail gen, waveform extraction, asset pool UI | 2 weeks |
| M2: Timeline MVP | Single video + audio track, clip drag/trim/split, playhead | 3 weeks |
| M3: Preview Player | Frame extraction, play/pause, scrubbing, canvas overlay | 2 weeks |
| M4: Basic Editing | Speed, opacity, crop, basic color (brightness/contrast/saturation) | 2 weeks |
| M5: Text & Export | Text overlay, FFmpeg filter graph, MP4 export with progress | 2 weeks |
| **v0.1 Release** | Public alpha | — |
| M6: Color Suite | Full color grading, LUT, filter presets, chroma key | 2 weeks |
| M7: Audio Engine | Waveform UI, fade handles, keyframes, noise reduction | 2 weeks |
| M8: Transitions | xfade transitions, transition panel, UI | 1 week |
| M9: Stickers & Keyframes | Sticker track, GIF support, keyframe animation system | 3 weeks |
| M10: Polish | Render queue, all export formats, performance profiling, cross-platform testing | 3 weeks |
| **v1.0 Release** | Stable public release | — |

**Total estimated v1.0:** ~6 months (solo developer, part-time)

---

*Gocut— Build fast. Edit local. Ship open.*
