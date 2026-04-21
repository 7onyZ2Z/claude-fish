package theme

import (
	"fmt"
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
func (claudeCodeTheme) UsableHeight(h int) int      { return h - 8 }

// renderClaudeHeader renders the shared header bar with branding, file info, and tips.
func renderClaudeHeader(version, fileName, themeName string, totalChapters int, width int) string {
	var b strings.Builder

	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudePurple))

	// Top border with branding
	topInner := fmt.Sprintf(" claude-fish %s ", version)
	rightInfo := fmt.Sprintf(" %s · %d ch ", fileName, totalChapters)
	padLen := width - 2 - len(topInner) - len(rightInfo)
	if padLen < 1 {
		padLen = 1
	}
	b.WriteString(borderStyle.Render("╭" + strings.Repeat("─", 1+len(topInner)) + " " + rightInfo + strings.Repeat("─", padLen) + "╮"))
	b.WriteString("\n")

	// Content row: logo left, tips right
	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeOrange))
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple))
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGray))

	logoLines := []string{
		fmt.Sprintf("│  %s  claude-fish %s", accentStyle.Render("●"), version),
		fmt.Sprintf("│          %s", orangeStyle.Render("▐▛███▜▌")),
		fmt.Sprintf("│         %s", orangeStyle.Render("▝▜█████▛▘")),
		fmt.Sprintf("│           %s", orangeStyle.Render("▘▘ ▝▝")),
		fmt.Sprintf("│  %s · %s", accentStyle.Render(themeName), grayStyle.Render(fileName)),
	}

	tipsLines := []string{
		grayStyle.Render("Tips for getting started"),
		grayStyle.Render(fmt.Sprintf("Loaded %s (%d chapters)", fileName, totalChapters)),
		grayStyle.Render("Space: next page | B: back"),
		grayStyle.Render("S: switch style | Tab: boss key"),
		grayStyle.Render("Q: quit"),
	}

	maxLines := len(logoLines)
	if len(tipsLines) > maxLines {
		maxLines = len(tipsLines)
	}

	for i := 0; i < maxLines; i++ {
		var left, right string
		if i < len(logoLines) {
			left = logoLines[i]
		}
		if i < len(tipsLines) {
			right = tipsLines[i]
		}
		// Pad left to 44 chars, right fills the rest
		leftPadded := fmt.Sprintf("%-44s", left)
		rightPad := width - 2 - 44
		if rightPad < 1 {
			rightPad = 1
		}
		rightPadded := fmt.Sprintf("%-*s", rightPad, right)
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

	// Code header bar
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudePurple)).
		Border(lipgloss.RoundedBorder(), true, false, true, false).
		BorderForeground(lipgloss.Color(claudePurple)).
		Width(width - 2)

	headerText := fmt.Sprintf(" ● Editing │ %s │ claude-fish", info.FileName)
	b.WriteString(headerStyle.Render(headerText))
	b.WriteString("\n")

	// AI preamble
	preambleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple))
	b.WriteString("✦ ")
	b.WriteString(preambleStyle.Render(fmt.Sprintf("Let me implement the changes in %s:", info.FileName)))
	b.WriteString("\n")

	// Code block
	if info.Displayed > 0 {
		codeContent := info.Content
		if info.Displayed < len(info.Content) {
			codeContent = info.Content[:info.Displayed]
		}

		codeStyle := lipgloss.NewStyle().Width(width - 4)
		fileLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGray)).Render(
			fmt.Sprintf("┌─ %s", info.FileName))
		b.WriteString(fileLabel)
		b.WriteString("\n")
		b.WriteString(codeStyle.Render(codeContent))

		if info.Displayed < info.Total {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(claudePurple)).Render("▌"))
		}
		b.WriteString("\n")
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
