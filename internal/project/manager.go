package project

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Manager struct {
	store *Store
}

func NewManager(store *Store) *Manager {
	return &Manager{store: store}
}

// NewProject creates a new project, applying the caller-supplied settings.
// The Go side stores name/aspectRatio/resolution/fps in addition to settings
// so the frontend can rely on a fully-populated project being returned.
func (m *Manager) NewProject(settings ProjectSettings) (*Project, error) {
	res := Resolution{Width: 1920, Height: 1080}
	fps := 30.0
	aspect := "16:9"
	name := "Untitled"

	if settings.Name != "" {
		name = settings.Name
	}
	if settings.AspectRatio != "" {
		aspect = settings.AspectRatio
	}
	if settings.Resolution != nil {
		res = *settings.Resolution
	}
	if settings.FPS > 0 {
		fps = settings.FPS
	}

	now := time.Now()
	p := &Project{
		ID:          generateID(),
		Name:        name,
		CreatedAt:   now,
		UpdatedAt:   now,
		Duration:    0,
		AspectRatio: aspect,
		Resolution:  res,
		FPS:         fps,
		Assets:      []Asset{},
		Timeline: Timeline{
			Tracks:   []Track{},
			Duration: 0,
		},
		Settings: ProjectSettings{
			BackgroundColor:  "#000000",
			AutoSave:         true,
			AutoSaveInterval: 60,
		},
	}

	// Seed default tracks so the editor UI has something to show.
	p.Timeline.Tracks = []Track{
		{ID: generateID(), Type: TrackVideo, Clips: []Clip{}, Muted: false, Locked: false},
		{ID: generateID(), Type: TrackAudio, Clips: []Clip{}, Muted: false, Locked: false, Volume: 1.0},
		{ID: generateID(), Type: TrackText, Clips: []Clip{}, Muted: false, Locked: false},
	}

	return p, nil
}

func (m *Manager) SaveProject(p Project) error {
	if m == nil || m.store == nil {
		return fmt.Errorf("project store is not initialized")
	}
	p.UpdatedAt = time.Now()

	if p.FilePath != "" {
		projDir := filepath.Dir(p.FilePath)
		// Make a shallow copy of the project to serialize relative paths to disk
		fileProj := p
		fileProj.Assets = make([]Asset, len(p.Assets))
		for i, asset := range p.Assets {
			fileProj.Assets[i] = asset
			if filepath.IsAbs(asset.Path) {
				relPath, err := filepath.Rel(projDir, asset.Path)
				if err == nil {
					fileProj.Assets[i].Path = relPath
				}
			}
		}

		data, err := json.MarshalIndent(fileProj, "", "  ")
		if err == nil {
			_ = os.WriteFile(p.FilePath, data, 0644)
		}
	}

	return m.store.SaveProject(p)
}

func (m *Manager) LoadProject(id string) (*Project, error) {
	if m == nil || m.store == nil {
		return nil, fmt.Errorf("project store is not initialized")
	}

	// Support opening exported project files directly from the file picker.
	if id != "" {
		if info, err := os.Stat(id); err == nil && !info.IsDir() {
			data, readErr := os.ReadFile(id)
			if readErr != nil {
				return nil, readErr
			}

			var p Project
			if err := json.Unmarshal(data, &p); err != nil {
				return nil, err
			}
			p.FilePath = id

			projDir := filepath.Dir(id)
			for i, asset := range p.Assets {
				if !filepath.IsAbs(asset.Path) {
					p.Assets[i].Path = filepath.Clean(filepath.Join(projDir, asset.Path))
				}
			}

			return &p, nil
		}
	}

	return m.store.LoadProject(id)
}

func (m *Manager) GetRecentProjects(limit int) ([]RecentProject, error) {
	if m == nil || m.store == nil {
		return []RecentProject{}, nil
	}
	return m.store.ListRecent(limit)
}

func (m *Manager) ExportProjectFile(p Project) (string, error) {
	filename := p.Name + ".Gocut"
	if filename == ".Gocut" {
		filename = "Untitled.Gocut"
	}
	path := filepath.Join(".", filename)

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func generateID() string {
	now := time.Now().UnixNano()
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%d_%s", now, hex.EncodeToString(b))
}
