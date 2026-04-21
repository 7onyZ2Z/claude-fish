package theme

import "github.com/charmbracelet/lipgloss"

type PageInfo struct {
	ChapterTitle  string
	Content       string
	PageNum       int
	TotalPages    int
	FileName      string
	TotalChapters int
	ThemeName     string
	Version       string
}

type CodeInfo struct {
	FileName     string
	Content      string
	Displayed    int
	Total        int
	ThemeName    string
	Version      string
}

type KeyHint struct {
	Key  string
	Desc string
}

type StatusInfo struct {
	Mode  string // "reading", "boss"
	Hints []KeyHint
}

type Theme interface {
	Name() string
	RenderPage(info PageInfo, width, height int) string
	RenderCode(info CodeInfo, width, height int) string
	RenderStatusBar(info StatusInfo, width int) string
	AccentColor() lipgloss.Color
	UsableHeight(termHeight int) int
}

// All returns all available themes in order.
func All() []Theme {
	return []Theme{
		NewClaudeCode(),
		NewCodex(),
		NewOpenCode(),
	}
}

// FindByName returns the theme with the given name, or the first theme if not found.
func FindByName(name string) Theme {
	for _, t := range All() {
		if t.Name() == name {
			return t
		}
	}
	return All()[0]
}
