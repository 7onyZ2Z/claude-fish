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

func NewClaudeCode() Theme {
	return claudeCodeTheme{}
}

func (claudeCodeTheme) Name() string { return "claude" }

func (claudeCodeTheme) AccentColor() lipgloss.Color { return claudePurple }

func (claudeCodeTheme) UsableHeight(termHeight int) int {
	return termHeight - 9
}

func (claudeCodeTheme) RenderWelcome(info WelcomeInfo) string {
	var b strings.Builder
	w := 80

	headerBorder := lipgloss.NewStyle().Foreground(lipgloss.Color(claudePurple))
	sepLine := strings.Repeat("─", w-22-len(info.Version))
	if len(sepLine) < 0 {
		sepLine = ""
	}
	b.WriteString(headerBorder.Render(fmt.Sprintf("╭─── claude-fish %s %s╮", info.Version, sepLine)))
	b.WriteString("\n")

	logoLines := []string{
		"                      ▐▛███▜▌",
		"                     ▝▜█████▛▘",
		"                       ▘▘ ▝▝",
	}

	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeOrange))
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple))

	tips := []string{
		fmt.Sprintf("Loaded %s", info.FileName),
		fmt.Sprintf("%d chapters found", info.Chapters),
		"Press Space to start reading",
		"Press Tab for boss mode",
	}

	leftLines := []string{
		"                  Welcome!",
		"",
	}
	leftLines = append(leftLines, logoLines...)
	leftLines = append(leftLines, "",
		fmt.Sprintf("   %s · %s", accentStyle.Render(info.ThemeName), info.Version),
	)

	maxLines := len(leftLines)
	if len(tips) > maxLines {
		maxLines = len(tips)
	}

	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGray))
	for i := 0; i < maxLines; i++ {
		var left, right string
		if i < len(leftLines) {
			left = leftLines[i]
		}
		if i < len(tips) {
			right = tips[i]
		}

		leftStyled := left
		if strings.Contains(left, "▐") || strings.Contains(left, "▜") || strings.Contains(left, "▝") {
			leftStyled = orangeStyle.Render(left)
		}

		leftPadded := fmt.Sprintf("%-44s", leftStyled)
		rightStyled := grayStyle.Render(right)
		b.WriteString(headerBorder.Render("│") + leftPadded + "│ " + rightStyled + "\n")
	}

	b.WriteString(headerBorder.Render(fmt.Sprintf("╰%s╯", strings.Repeat("─", w-2))))
	b.WriteString("\n")
	return b.String()
}

func (claudeCodeTheme) RenderPage(info PageInfo, width, height int) string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudePurple)).
		Border(lipgloss.RoundedBorder(), true, false, true, false).
		BorderForeground(lipgloss.Color(claudePurple)).
		Width(width - 2)

	headerText := fmt.Sprintf(" ● Reading %s │ %s │ Page %d/%d",
		info.FileName, info.ChapterTitle, info.PageNum+1, info.TotalPages)
	b.WriteString(headerStyle.Render(headerText))
	b.WriteString("\n")

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudeLightPurple)).
		Bold(true)
	b.WriteString("✦ ")
	b.WriteString(titleStyle.Render(info.ChapterTitle))
	b.WriteString("\n")

	bubbleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudeLightGray)).
		Background(lipgloss.Color(claudeDarkPurple)).
		BorderLeft(true).
		BorderBackground(lipgloss.Color(claudeDarkPurple)).
		BorderForeground(lipgloss.Color(claudePurple)).
		Width(width - 4)

	b.WriteString(bubbleStyle.Render(info.Content))
	b.WriteString("\n")

	toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGreen))
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGray))
	b.WriteString(toolStyle.Render("✔"))
	b.WriteString(grayStyle.Render(fmt.Sprintf(" Page %d of %d", info.PageNum+1, info.TotalPages)))
	b.WriteString("\n")

	return b.String()
}

func (claudeCodeTheme) RenderCode(info CodeInfo, width, height int) string {
	var b strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudePurple)).
		Border(lipgloss.RoundedBorder(), true, false, true, false).
		BorderForeground(lipgloss.Color(claudePurple)).
		Width(width - 2)

	headerText := fmt.Sprintf(" ● Editing %s │ claude-fish", info.FileName)
	b.WriteString(headerStyle.Render(headerText))
	b.WriteString("\n")

	preambleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple))
	b.WriteString("✦ ")
	b.WriteString(preambleStyle.Render(fmt.Sprintf("Let me implement the changes in %s:", info.FileName)))
	b.WriteString("\n")

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
	case "welcome":
		parts = []string{"Space: start", "Q: quit"}
	case "reading":
		parts = []string{"Space: next", "B: back", "S: style", "Tab: boss", "Q: quit"}
	case "boss":
		parts = []string{"Tab: back to novel"}
	}

	var styled []string
	for _, p := range parts {
		// Split on ": " to style the key differently
		idx := strings.Index(p, ": ")
		if idx >= 0 {
			styled = append(styled, hintStyle.Render(p[:idx])+": "+p[idx+2:])
		} else {
			styled = append(styled, p)
		}
	}

	return strings.Join(styled, sepChar)
}
