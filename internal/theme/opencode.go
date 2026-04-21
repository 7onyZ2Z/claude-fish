package theme

import "github.com/charmbracelet/lipgloss"

type opencodeTheme struct{}

func NewOpenCode() Theme { return opencodeTheme{} }
func (opencodeTheme) Name() string                          { return "opencode" }
func (opencodeTheme) AccentColor() lipgloss.Color           { return "#89b4fa" }
func (opencodeTheme) UsableHeight(h int) int                { return h - 7 }
func (opencodeTheme) RenderWelcome(WelcomeInfo) string       { return "welcome\n" }
func (opencodeTheme) RenderPage(PageInfo, int, int) string   { return "page\n" }
func (opencodeTheme) RenderCode(CodeInfo, int, int) string   { return "code\n" }
func (opencodeTheme) RenderStatusBar(StatusInfo, int) string { return "status\n" }
