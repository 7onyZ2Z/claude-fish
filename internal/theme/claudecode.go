package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	claudePurple      = "#7c3aed"
	claudeLightPurple = "#a78bfa"
	claudeOrange      = "#f97316"
	claudeGreen       = "#22c55e"
	claudeDarkPurple  = "#2d1b69"
	claudeDarkBg      = "#1a1a2e"
	claudeGray        = "#6b7280"
	claudeLightGray   = "#d1d5db"
)

type claudeCodeTheme struct{}

func NewClaudeCode() Theme { return claudeCodeTheme{} }

func (claudeCodeTheme) Name() string                { return "claude" }
func (claudeCodeTheme) AccentColor() lipgloss.Color { return claudePurple }
func (claudeCodeTheme) UsableHeight(h int) int {
	// header: 11 lines, reading bar: 3, title: 1, page indicator: 1
	// view overhead: separator + input + separator = 3
	// total fixed: 19, leave 1 line margin
	return h - 20
}

// padRight pads a string to target visible width (accounts for ANSI escape codes).
func padRight(s string, targetWidth int) string {
	visWidth := lipgloss.Width(s)
	if visWidth >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-visWidth)
}

// renderClaudeHeader renders a realistic Claude Code header for disguise.
func renderClaudeHeader(version, fileName, themeName string, totalChapters int, width int) string {
	var b strings.Builder

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudePurple))
	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeOrange))
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple))
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGray))

	// Top border
	topText := fmt.Sprintf(" Claude Code %s ", version)
	topVisibleW := lipgloss.Width(topText)
	rightText := " Tips for getting started "
	rightVisibleW := lipgloss.Width(rightText)
	padLen := width - 2 - topVisibleW - rightVisibleW
	if padLen < 1 {
		padLen = 1
	}
	b.WriteString(borderStyle.Render("╭" + topText + rightText + strings.Repeat("─", padLen) + "╮"))
	b.WriteString("\n")

	// Working directory
	cwd, _ := os.Getwd()
	shortDir := filepath.Base(cwd)
	if cwd != "" {
		shortDir = "~/" + shortDir
	}

	innerW := width - 2
	leftColW := innerW / 2

	type row struct{ left, right string }
	rows := []row{
		{"", ""},
		{"                  Welcome back!", ""},
		{"", "Run /init to create a CLAUDE.md file with instructions for Claude"},
		{"                      " + orangeStyle.Render("▐▛███▜▌"), " ─────────────────────────────────────────────────────────────"},
		{"                     " + orangeStyle.Render("▝▜█████▛▘"), " Recent activity"},
		{"                       " + orangeStyle.Render("▘▘ ▝▝"), " No recent activity"},
		{"", ""},
		{"   " + accentStyle.Render("Claude Opus 4.6 with medium effort · API Usage Billing"), ""},
		{"                " + grayStyle.Render(shortDir), ""},
	}

	for _, r := range rows {
		leftPadded := padRight(r.left, leftColW)
		rightPadded := padRight(r.right, innerW-leftColW)
		b.WriteString(borderStyle.Render("│") + leftPadded + rightPadded + borderStyle.Render("│"))
		b.WriteString("\n")
	}

	// Bottom border
	b.WriteString(borderStyle.Render(fmt.Sprintf("╰%s╯", strings.Repeat("─", width-2))))
	b.WriteString("\n")

	return b.String()
}

func (claudeCodeTheme) RenderPage(info PageInfo, width, height int) string {
	var b strings.Builder

	// Shared header
	b.WriteString(renderClaudeHeader(info.Version, info.FileName, info.ThemeName, info.TotalChapters, width))

	// Content header bar
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudePurple)).
		Border(lipgloss.RoundedBorder(), true, false, true, false).
		BorderForeground(lipgloss.Color(claudePurple)).
		Width(width - 2)

	headerText := fmt.Sprintf(" ● Reading │ %s │ Page %d/%d",
		info.ChapterTitle, info.PageNum+1, info.TotalPages)
	b.WriteString(headerStyle.Render(headerText))
	b.WriteString("\n")

	// Chapter title
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple)).Bold(true)
	b.WriteString("✦ ")
	b.WriteString(titleStyle.Render(info.ChapterTitle))
	b.WriteString("\n")

	// Content bubble
	bubbleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudeLightGray)).
		Background(lipgloss.Color(claudeDarkPurple)).
		BorderLeft(true).
		BorderBackground(lipgloss.Color(claudeDarkPurple)).
		BorderForeground(lipgloss.Color(claudePurple)).
		Width(width - 4)
	b.WriteString(bubbleStyle.Render(info.Content))
	b.WriteString("\n")

	// Page indicator
	toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGreen))
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGray))
	b.WriteString(toolStyle.Render("✔"))
	b.WriteString(grayStyle.Render(fmt.Sprintf(" Page %d of %d", info.PageNum+1, info.TotalPages)))
	b.WriteString("\n")

	return b.String()
}

func (claudeCodeTheme) RenderCode(info CodeInfo, width, height int) string {
	var b strings.Builder

	// Shared header
	b.WriteString(renderClaudeHeader(info.Version, info.FileName, info.ThemeName, 0, width))

	// Content — already formatted with think/text/code segments by streamer
	if info.Content != "" {
		b.WriteString(info.Content)
		if !strings.HasSuffix(info.Content, "\n") {
			b.WriteString("\n")
		}
		if info.Displayed < info.Total {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(claudePurple)).Render("▌"))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (claudeCodeTheme) RenderStatusBar(info StatusInfo, width int) string {
	sepChar := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563")).Render(" | ")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple))

	var parts []string
	switch info.Mode {
	case "reading":
		parts = []string{"Space: next", "B: back", "S: style", "Tab: boss", "Q: quit"}
	case "boss":
		parts = []string{"Tab: back to novel"}
	}

	var styled []string
	for _, p := range parts {
		idx := strings.Index(p, ": ")
		if idx >= 0 {
			styled = append(styled, hintStyle.Render(p[:idx])+": "+p[idx+2:])
		} else {
			styled = append(styled, p)
		}
	}

	return strings.Join(styled, sepChar)
}
