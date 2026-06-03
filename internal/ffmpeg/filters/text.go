package filters

import (
	"fmt"
	"strings"

	"Gocut/internal/project"
)

func BuildTextFilter(p project.TextProps) string {
	if p.Text == "" {
		return ""
	}
	filter := "drawtext=text='" + escapeText(p.Text) + "'"
	if p.FontFamily != "" {
		filter += ":fontfile='" + p.FontFamily + "'"
	}
	filter += fmt.Sprintf(":fontsize=%d", p.FontSize)
	if p.Color != "" {
		filter += ":fontcolor=" + p.Color
	}
	return filter
}

func escapeText(s string) string {
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, ":", "\\:")
	return s
}
