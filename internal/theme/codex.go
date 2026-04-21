package theme

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	codexGreen = "#10a37f"
	codexGray  = "#6b7280"
	codexWhite = "#e0e0e0"
)

type codexTheme struct{}

func NewCodex() Theme { return codexTheme{} }

func (codexTheme) Name() string                { return "codex" }
func (codexTheme) AccentColor() lipgloss.Color { return codexGreen }
func (codexTheme) UsableHeight(h int) int      { return h - 5 }

// renderCodexHeader renders the shared header for reading and editing views.
func renderCodexHeader(version, fileName, themeName string, totalChapters int) string {
	var b strings.Builder
	green := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGreen)).Bold(true)
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGray))

	b.WriteString(green.Render("codex"))
	b.WriteString(" ")
	b.WriteString(gray.Render(version))
	b.WriteString(gray.Render(" │ "))
	b.WriteString(gray.Render(fmt.Sprintf("%s (%d ch) │ %s", fileName, totalChapters, themeName)))
	b.WriteString("\n")
	b.WriteString(gray.Render(strings.Repeat("─", 60)))
	b.WriteString("\n")

	return b.String()
}

func (codexTheme) RenderPage(info PageInfo, width, _ int) string {
	var b strings.Builder

	// Shared header
	b.WriteString(renderCodexHeader(info.Version, info.FileName, info.ThemeName, info.TotalChapters))

	green := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGreen))
	white := lipgloss.NewStyle().Foreground(lipgloss.Color(codexWhite))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGray))

	b.WriteString(green.Render(">"))
	b.WriteString(" ")
	b.WriteString(white.Render(info.ChapterTitle))
	b.WriteString("\n\n")

	b.WriteString(white.Render(info.Content))
	b.WriteString("\n\n")

	progress := float64(info.PageNum+1) / float64(info.TotalPages)
	barWidth := 30
	filled := int(progress * float64(barWidth))
	bar := green.Render(strings.Repeat("█", filled)) + gray.Render(strings.Repeat("░", barWidth-filled))
	b.WriteString(fmt.Sprintf("[%s] %d/%d", bar, info.PageNum+1, info.TotalPages))
	b.WriteString("\n")

	return b.String()
}

func (codexTheme) RenderCode(info CodeInfo, width, _ int) string {
	var b strings.Builder

	// Shared header
	b.WriteString(renderCodexHeader(info.Version, info.FileName, info.ThemeName, 0))

	green := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGreen))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGray))
	white := lipgloss.NewStyle().Foreground(lipgloss.Color(codexWhite))

	b.WriteString(green.Render(">"))
	b.WriteString(" ")
	b.WriteString(white.Render(fmt.Sprintf("Editing %s", info.FileName)))
	b.WriteString("\n\n")

	if info.Displayed > 0 {
		visible := info.Content
		if info.Displayed < len(info.Content) {
			visible = info.Content[:info.Displayed]
		}
		b.WriteString(gray.Render(fmt.Sprintf("┌─ %s", info.FileName)))
		b.WriteString("\n")
		b.WriteString(white.Render(visible))
		if info.Displayed < info.Total {
			b.WriteString(green.Render("▌"))
		}
		b.WriteString("\n")
	}

	b.WriteString(gray.Render("Writing..."))
	b.WriteString("\n")

	return b.String()
}

func (codexTheme) RenderStatusBar(info StatusInfo, width int) string {
	var parts []string
	switch info.Mode {
	case "reading":
		parts = []string{"Space: next", "B: back", "S: style", "!: boss", "Q: quit"}
	case "boss":
		parts = []string{"Tab: back"}
	}

	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563")).Render(" | ")
	green := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGreen))

	return green.Render("→") + sep + strings.Join(parts, sep)
}
