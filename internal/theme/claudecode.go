package theme

import "github.com/charmbracelet/lipgloss"

type claudeCodeTheme struct{}

func NewClaudeCode() Theme { return claudeCodeTheme{} }
func (claudeCodeTheme) Name() string                          { return "claude" }
func (claudeCodeTheme) AccentColor() lipgloss.Color           { return "#7c3aed" }
func (claudeCodeTheme) UsableHeight(h int) int                { return h - 9 }
func (claudeCodeTheme) RenderWelcome(WelcomeInfo) string       { return "welcome\n" }
func (claudeCodeTheme) RenderPage(PageInfo, int, int) string   { return "page\n" }
func (claudeCodeTheme) RenderCode(CodeInfo, int, int) string   { return "code\n" }
func (claudeCodeTheme) RenderStatusBar(StatusInfo, int) string { return "status\n" }
