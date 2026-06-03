package fonts

import (
	"os"
	"path/filepath"
	"strings"
)

type Scanner struct{}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Scan() []FontInfo {
	var fonts []FontInfo

	dirs := fontDirs()
	for _, dir := range dirs {
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".ttf" || ext == ".otf" || ext == ".ttc" {
				fonts = append(fonts, FontInfo{
					Family: filepath.Base(path),
					Path:   path,
					Style:  "",
				})
			}
			return nil
		})
	}

	return fonts
}

func fontDirs() []string {
	var dirs []string
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "Library", "Fonts"))
		dirs = append(dirs, filepath.Join(home, ".fonts"))
	}
	dirs = append(dirs, "/usr/share/fonts")
	dirs = append(dirs, "/usr/local/share/fonts")
	dirs = append(dirs, "C:\\Windows\\Fonts")
	return dirs
}

type FontInfo struct {
	Family string `json:"family"`
	Path   string `json:"path"`
	Style  string `json:"style"`
}
