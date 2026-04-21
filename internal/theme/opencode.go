package theme

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	ocBlue    = "#89b4fa"
	ocText    = "#cdd6f4"
	ocSubtext = "#6c7086"
	ocOverlay = "#45475a"
	ocRed     = "#f38ba8"
)

type opencodeTheme struct{}

func NewOpenCode() Theme { return opencodeTheme{} }

func (opencodeTheme) Name() string                { return "opencode" }
func (opencodeTheme) AccentColor() lipgloss.Color { return ocBlue }
func (opencodeTheme) UsableHeight(h int) int      { return h - 7 }

func (opencodeTheme) RenderWelcome(info WelcomeInfo) string {
	var b strings.Builder
	blue := lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue)).Bold(true)
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))

	b.WriteString(renderTabBar("Welcome", []string{"Welcome", "Files", "Config"}))
	b.WriteString("\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ocOverlay)).
		Width(60).
		Padding(1)

	content := blue.Render("opencode") + "\n\n" +
		fmt.Sprintf("Loaded: %s (%d chapters)\n", info.FileName, info.Chapters) +
		gray.Render("Press Space to start")
	b.WriteString(box.Render(content))
	b.WriteString("\n")
	return b.String()
}

func (opencodeTheme) RenderPage(info PageInfo, width, _ int) string {
	var b strings.Builder

	b.WriteString(renderTabBar("Chat", []string{"Chat", "Files", "Diff"}))
	b.WriteString("\n")

	blue := lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue))
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(ocText))

	b.WriteString(blue.Render("assistant"))
	b.WriteString("\n")
	b.WriteString(text.Render("  " + info.ChapterTitle))
	b.WriteString("\n\n")
	b.WriteString(text.Render(info.Content))
	b.WriteString("\n\n")

	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))
	b.WriteString(gray.Render(fmt.Sprintf("%d/%d pages", info.PageNum+1, info.TotalPages)))
	b.WriteString("\n")

	return b.String()
}

func (opencodeTheme) RenderCode(info CodeInfo, width, _ int) string {
	var b strings.Builder

	b.WriteString(renderTabBar("Diff", []string{"Chat", "Files", "Diff"}))
	b.WriteString("\n")

	blue := lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue))
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(ocText))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))

	b.WriteString(blue.Render("assistant"))
	b.WriteString("\n")

	if info.Displayed > 0 {
		visible := info.Content
		if info.Displayed < len(info.Content) {
			visible = info.Content[:info.Displayed]
		}
		b.WriteString(gray.Render(fmt.Sprintf("┌─ %s", info.FileName)))
		b.WriteString("\n")
		b.WriteString(text.Render(visible))
		if info.Displayed < info.Total {
			b.WriteString(blue.Render("▌"))
		}
		b.WriteString("\n")
	}

	b.WriteString(gray.Render("Writing..."))
	b.WriteString("\n")
	return b.String()
}

func (opencodeTheme) RenderStatusBar(info StatusInfo, width int) string {
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color(ocRed))

	switch info.Mode {
	case "welcome":
		return gray.Render("Space: start | q: quit")
	case "reading":
		return gray.Render("Space: next | B: back | S: style | ") + red.Render("Tab") + gray.Render(": boss")
	case "boss":
		return red.Render("[BOSS MODE] ") + gray.Render("Tab: back")
	}
	return ""
}

func renderTabBar(active string, tabs []string) string {
	var b strings.Builder
	blue := lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))

	for _, tab := range tabs {
		if tab == active {
			b.WriteString(blue.Bold(true).Render(tab))
			b.WriteString("  ")
		} else {
			b.WriteString(gray.Render(tab))
			b.WriteString("  ")
		}
	}
	b.WriteString("\n")
	for _, tab := range tabs {
		if tab == active {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue)).Render(strings.Repeat("─", len(tab))))
			b.WriteString("  ")
		} else {
			b.WriteString(strings.Repeat(" ", len(tab)))
			b.WriteString("  ")
		}
	}
	return b.String()
}
