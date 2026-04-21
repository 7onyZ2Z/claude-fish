package theme

import "github.com/charmbracelet/lipgloss"

type codexTheme struct{}

func NewCodex() Theme { return codexTheme{} }
func (codexTheme) Name() string                          { return "codex" }
func (codexTheme) AccentColor() lipgloss.Color           { return "#10a37f" }
func (codexTheme) UsableHeight(h int) int                { return h - 6 }
func (codexTheme) RenderWelcome(WelcomeInfo) string       { return "welcome\n" }
func (codexTheme) RenderPage(PageInfo, int, int) string   { return "page\n" }
func (codexTheme) RenderCode(CodeInfo, int, int) string   { return "code\n" }
func (codexTheme) RenderStatusBar(StatusInfo, int) string { return "status\n" }
