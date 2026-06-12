package project

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
)

type Store struct {
	db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	schema := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TEXT,
			updated_at TEXT,
			duration REAL,
			aspect_ratio TEXT,
			resolution_w INTEGER,
			resolution_h INTEGER,
			fps REAL,
			data TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS assets (
			id TEXT PRIMARY KEY,
			project_id TEXT,
			path TEXT,
			type TEXT,
			duration REAL,
			width INTEGER,
			height INTEGER,
			fps REAL,
			codec TEXT,
			thumbnail TEXT,
			waveform TEXT,
			file_size INTEGER,
			imported_at TEXT
		)`,
	}

	for _, stmt := range schema {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SaveProject(p Project) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO projects (id, name, created_at, updated_at, duration, aspect_ratio, resolution_w, resolution_h, fps, data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Name, p.CreatedAt.Format(time.RFC3339), p.UpdatedAt.Format(time.RFC3339),
		p.Duration, p.AspectRatio, p.Resolution.Width, p.Resolution.Height, p.FPS, string(data),
	)
	return err
}

func (s *Store) LoadProject(id string) (*Project, error) {
	row := s.db.QueryRow(`SELECT data FROM projects WHERE id = ?`, id)

	var data string
	if err := row.Scan(&data); err != nil {
		return nil, err
	}

	var p Project
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Store) ListRecent(limit int) ([]RecentProject, error) {
	rows, err := s.db.Query(`
		SELECT id, name, updated_at, data FROM projects
		ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []RecentProject
	for rows.Next() {
		var r RecentProject
		var updated string
		var data string
		var dbID string
		if err := rows.Scan(&dbID, &r.Name, &updated, &data); err != nil {
			continue
		}
		r.ID = dbID
		r.Path = dbID
		r.UpdatedAt, _ = time.Parse(time.RFC3339, updated)

		var proj Project
		if err := json.Unmarshal([]byte(data), &proj); err == nil && proj.FilePath != "" {
			r.Path = proj.FilePath
		}

		list = append(list, r)
	}
	return list, nil
}

func (s *Store) DeleteProject(id string) error {
	_, err := s.db.Exec(`DELETE FROM projects WHERE id = ?`, id)
	return err
}

func (s *Store) ClearRecent() error {
	_, err := s.db.Exec(`DELETE FROM projects`)
	return err
}

func (s *Store) Close() error {
	return s.db.Close()
}
