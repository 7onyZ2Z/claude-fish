package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"claude-fish/internal/reader"
	"claude-fish/internal/theme"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── States ───────────────────────────────────────────────────────

type appState int

const (
	stateReading appState = iota
	stateBoss
)

type menuKind int

const (
	menuNone menuKind = iota
	menuNovel
	menuTheme
	menuChapter
	menuResume
)

type menuItem struct {
	label string
	value string
}

type menuState struct {
	kind     menuKind
	items    []menuItem
	selected int
}

// ── Command registry ─────────────────────────────────────────────

type command struct {
	name string
	desc string
}

var allCommands = []command{
	{"/novel", "Switch novel file"},
	{"/theme", "Switch theme style"},
	{"/chapters", "Jump to chapter"},
	{"/back", "Previous page"},
	{"/next", "Next page"},
}

// ── Model ────────────────────────────────────────────────────────

type model struct {
	state      appState
	theme      theme.Theme
	themes     []theme.Theme
	themeIndex int
	pager      *Pager
	boss       *BossMode
	width      int
	height     int
	speed      int
	fileName   string
	version    string
	input      textinput.Model
	menu       menuState
	novelDir   string
	history    *HistoryStore
	resumePath string // pending novel path awaiting resume choice
}

func newModel(r reader.Reader, th theme.Theme, boss *BossMode, width, height, speed int) model {
	themes := theme.All()
	themeIndex := 0
	for i, t := range themes {
		if t.Name() == th.Name() {
			themeIndex = i
			break
		}
	}

	var pg *Pager
	if r != nil {
		usableHeight := th.UsableHeight(height)
		if usableHeight < 1 {
			usableHeight = 1
		}
		pg = NewPager(r, width-4, usableHeight)
	}

	if boss == nil {
		boss = NewBossMode(defaultBossContent, speed)
	}

	ti := textinput.New()
	ti.Prompt = "❯ "
	ti.Placeholder = "Type a message... or type / for commands"
	ti.CharLimit = 500
	ti.Width = width - 4

	return model{
		state:      stateReading,
		theme:      th,
		themes:     themes,
		themeIndex: themeIndex,
		pager:      pg,
		boss:       boss,
		width:      width,
		height:     height,
		speed:      speed,
		version:    "v2.1.116",
		input:      ti,
		novelDir:   "novel",
		history:    LoadHistory(),
	}
}

func NewApp(r reader.Reader, th theme.Theme, segments []Segment, fn string, speed int) *tea.Program {
	m := newModel(r, th, NewBossMode(segments, speed), 80, 24, speed)
	m.fileName = fn
	m.input.Focus()
	return tea.NewProgram(m, tea.WithAltScreen())
}

func (model) Init() tea.Cmd { return textinput.Blink }

type tickMsg time.Time

// ── Update ───────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 4
		if m.pager != nil {
			usable := m.theme.UsableHeight(msg.Height)
			if usable < 1 {
				usable = 1
			}
			m.pager.Resize(msg.Width-4, usable)
		}
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		// If a submenu is active (novel/theme/chapter picker), handle it
		if m.menu.kind != menuNone {
			return m.handleMenuKey(key)
		}

		// Tab: boss mode toggle (always a shortcut)
		if key == "tab" {
			return m.handleTab()
		}
		// Esc: close autocomplete if showing, otherwise boss back or quit
		if key == "esc" {
			input := m.input.Value()
			if strings.HasPrefix(input, "/") && len(filterCommands(input)) > 0 {
				m.input.SetValue("")
				m.menu = menuState{}
				return m, nil
			}
			return m, tea.Quit
		}

		// Enter: process command or show output
		if key == "enter" {
			return m.handleEnter()
		}

		// Space: page next only when input is empty
		if key == " " {
			if strings.TrimSpace(m.input.Value()) == "" {
				if m.state == stateReading && m.pager != nil {
					m.pager.NextPage()
					m.saveHistory()
				}
				return m, nil
			}
			// Input has content → space goes to input as normal char
		}

		// Up/down: if command autocomplete is showing, navigate it
		input := m.input.Value()
		if strings.HasPrefix(input, "/") {
			filtered := filterCommands(input)
			if key == "up" && len(filtered) > 0 {
				if m.menu.selected > 0 {
					m.menu.selected--
				}
				return m, nil
			}
			if key == "down" && len(filtered) > 0 {
				if m.menu.selected < len(filtered)-1 {
					m.menu.selected++
				}
				return m, nil
			}
		}

		// All other keys go to text input
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		// Reset autocomplete selection when input changes
		m.menu = menuState{}
		return m, cmd

	case tickMsg:
		if m.state == stateBoss && m.boss.HasCode() {
			s := m.boss.Streamer()
			if s.Done() {
				s.Reset()
			}
			s.Advance(s.ChunkSize())
			return m, tea.Tick(time.Duration(s.JitterSpeed())*time.Millisecond,
				func(t time.Time) tea.Msg { return tickMsg(t) })
		}
		return m, nil
	}

	return m, nil
}

// ── Tab ──────────────────────────────────────────────────────────

func (m model) handleTab() (tea.Model, tea.Cmd) {
	if m.state == stateReading && m.boss.HasCode() {
		m.state = stateBoss
		m.boss.Activate()
		return m, tea.Tick(time.Duration(m.boss.Streamer().JitterSpeed())*time.Millisecond,
			func(t time.Time) tea.Msg { return tickMsg(t) })
	} else if m.state == stateBoss {
		m.state = stateReading
		m.boss.Deactivate()
	}
	return m, nil
}

// ── Enter: command processing ────────────────────────────────────

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.input.Value())
	m.input.SetValue("")

	if input == "" {
		return m, nil
	}

	// If autocomplete is showing and something is selected, use that
	if strings.HasPrefix(input, "/") && m.menu.selected >= 0 {
		filtered := filterCommands(input)
		if m.menu.selected < len(filtered) {
			input = filtered[m.menu.selected].name
		}
	}
	m.menu = menuState{}

	if !strings.HasPrefix(input, "/") {
		// Not a command → show as user message, no action
		return m, nil
	}

	parts := strings.SplitN(input, " ", 2)
	cmd := parts[0]
	arg := ""
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}

	switch cmd {
	case "/novel":
		return m.cmdNovel()
	case "/theme":
		return m.cmdTheme()
	case "/chapters":
		return m.cmdChapters(arg)
	case "/back":
		if m.pager != nil {
			m.pager.PrevPage()
		}
	case "/next":
		if m.pager != nil {
			m.pager.NextPage()
		}
	case "/help":
		var lines []string
		for _, c := range allCommands {
			lines = append(lines, fmt.Sprintf("  %-12s %s", c.name, c.desc))
		}
	default:
	}

	return m, nil
}

// ── Command implementations ──────────────────────────────────────

func (m model) cmdNovel() (tea.Model, tea.Cmd) {
	items := scanNovelDir(m.novelDir)
	if len(items) == 0 {
		return m, nil
	}
	m.menu = menuState{kind: menuNovel, items: items, selected: 0}
	return m, nil
}

func (m model) cmdTheme() (tea.Model, tea.Cmd) {
	var items []menuItem
	for _, t := range m.themes {
		items = append(items, menuItem{label: t.Name(), value: t.Name()})
	}
	m.menu = menuState{kind: menuTheme, items: items, selected: m.themeIndex}
	return m, nil
}

func (m model) cmdChapters(arg string) (tea.Model, tea.Cmd) {
	if arg != "" {
		n, err := strconv.Atoi(arg)
		if err == nil && m.pager != nil {
			chapters := m.pager.Chapters()
			if n >= 1 && n <= len(chapters) {
				m.pager.GoToChapter(n - 1)
			} else {
			}
		}
		return m, nil
	}
	if m.pager == nil {
		return m, nil
	}
	chapters := m.pager.Chapters()
	var items []menuItem
	for i, ch := range chapters {
		items = append(items, menuItem{
			label: fmt.Sprintf("%d. %s", i+1, ch.Title),
			value: strconv.Itoa(i),
		})
	}
	m.menu = menuState{kind: menuChapter, items: items, selected: m.pager.Chapter()}
	return m, nil
}

// ── Menu key handling ────────────────────────────────────────────

func (m model) handleMenuKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "ctrl+p":
		if m.menu.selected > 0 {
			m.menu.selected--
		}
	case "down", "ctrl+n":
		if m.menu.selected < len(m.menu.items)-1 {
			m.menu.selected++
		}
	case "enter":
		return m.confirmMenuSelection()
	case "esc":
		m.menu = menuState{}
	}
	return m, nil
}

func (m model) confirmMenuSelection() (tea.Model, tea.Cmd) {
	if m.menu.selected < 0 || m.menu.selected >= len(m.menu.items) {
		m.menu = menuState{}
		return m, nil
	}

	item := m.menu.items[m.menu.selected]

	switch m.menu.kind {
	case menuNovel:
		path := item.value
		r := newReaderForFile(path)
		if r != nil {
			if err := r.Load(path); err == nil {
				usable := m.theme.UsableHeight(m.height)
				if usable < 1 {
					usable = 1
				}
				m.pager = NewPager(r, m.width-4, usable)
				m.fileName = filepath.Base(path)
				m.resumePath = path

				if pos, ok := m.history.Get(path); ok {
					chTitle := "(unknown chapter)"
					if pos.Chapter < len(m.pager.Chapters()) {
						chTitle = m.pager.Chapters()[pos.Chapter].Title
					}
					m.menu = menuState{kind: menuResume, selected: 0, items: []menuItem{
						{label: fmt.Sprintf("Resume: %s (page %d)", chTitle, pos.Page+1), value: "resume"},
						{label: "Start from beginning", value: "start"},
						{label: "Jump to chapter...", value: "jump"},
					}}
					return m, nil
				}
			}
		}
	case menuTheme:
		for i, t := range m.themes {
			if t.Name() == item.value {
				m.themeIndex = i
				m.theme = m.themes[i]
				if m.pager != nil {
					usable := m.theme.UsableHeight(m.height)
					if usable < 1 {
						usable = 1
					}
					m.pager.SetThemeLines(usable)
				}
				break
			}
		}
	case menuChapter:
		ch, _ := strconv.Atoi(item.value)
		if m.pager != nil {
			m.pager.GoToChapter(ch)
			m.saveHistory()
		}
	case menuResume:
		switch item.value {
		case "resume":
			if pos, ok := m.history.Get(m.resumePath); ok && m.pager != nil {
				m.pager.GoToChapter(pos.Chapter)
				m.pager.SetPage(pos.Page)
			}
		case "start":
		case "jump":
			m.menu = menuState{}
			return m.cmdChapters("")
		}
		m.resumePath = ""
	}

	m.menu = menuState{}
	m.saveHistory()
	return m, nil
}

func (m *model) saveHistory() {
	if m.pager != nil && m.fileName != "" && m.history != nil {
		m.history.Save(filepath.Join(m.novelDir, m.fileName), m.pager.Chapter(), m.pager.Page())
	}
}

// ── Command autocomplete ─────────────────────────────────────────

func filterCommands(input string) []command {
	var filtered []command
	for _, c := range allCommands {
		if strings.HasPrefix(c.name, input) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// ── File helpers ─────────────────────────────────────────────────

func scanNovelDir(dir string) []menuItem {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var supported = map[string]bool{".txt": true, ".md": true, ".markdown": true, ".epub": true}
	var items []menuItem

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if supported[ext] {
			fullPath := filepath.Join(dir, e.Name())
			items = append(items, menuItem{
				label: e.Name(),
				value: fullPath,
			})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].label < items[j].label
	})
	return items
}

func newReaderForFile(path string) reader.Reader {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt":
		return &reader.TXTReader{}
	case ".md", ".markdown":
		return &reader.MarkdownReader{}
	case ".epub":
		return &reader.EPUBReader{}
	}
	return nil
}

// ── View ─────────────────────────────────────────────────────────

func (m model) View() string {
	inputView := m.input.View()

	// 1. Content area
	var content string
	switch m.state {
	case stateReading:
		content = m.renderReadingContent()
	case stateBoss:
		s := m.boss.Streamer()
		visible := s.VisibleContent()
		content = m.theme.RenderCode(theme.CodeInfo{
			FileName:  "main.go",
			Content:   visible,
			Displayed: s.Displayed(),
			Total:     s.Total(),
			ThemeName: m.theme.Name(),
			Version:   m.version,
		}, m.width, m.height)
	}

	popupView := m.renderPopup()

	output := content + renderSeparator(m.width) + "\n" + inputView + "\n" + renderSeparator(m.width) + "\n" + popupView

	// Pad output to fill terminal so old content doesn't bleed through.
	// Leave 1 line margin to prevent top border from scrolling off screen.
	lines := strings.Count(output, "\n")
	target := m.height - 1
	if lines < target {
		output += strings.Repeat("\n", target-lines)
	}

	return output
}

func (m model) renderReadingContent() string {
	if m.pager == nil {
		hint := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6b7280")).
			Render("No novel loaded. Type /novel to select a file.")
		return m.theme.RenderPage(theme.PageInfo{
			Content:   hint,
			ThemeName: m.theme.Name(),
			Version:   m.version,
		}, m.width, m.height)
	}
	title := m.pager.CurrentTitle()
	content := m.pager.CurrentContent()
	pageNum := m.pager.Page()
	totalPages := m.pager.TotalPages()
	totalChapters := m.pager.TotalChapters()
	return m.theme.RenderPage(theme.PageInfo{
		ChapterTitle:  title,
		Content:       content,
		PageNum:       pageNum,
		TotalPages:    totalPages,
		FileName:      m.fileName,
		TotalChapters: totalChapters,
		ThemeName:     m.theme.Name(),
		Version:       m.version,
	}, m.width, m.height)
}

func (m model) renderPopup() string {
	// Submenu picker (novel/theme/chapter)
	if m.menu.kind != menuNone {
		return m.renderMenu()
	}

	// Autocomplete: show filtered commands when input starts with /
	input := m.input.Value()
	if !strings.HasPrefix(input, "/") {
		return ""
	}
	filtered := filterCommands(input)
	if len(filtered) == 0 {
		return ""
	}

	accentColor := m.theme.AccentColor()
	selectedStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#d1d5db"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))

	var b strings.Builder
	for i, cmd := range filtered {
		if i == m.menu.selected {
			b.WriteString(selectedStyle.Render(fmt.Sprintf("  > %-12s", cmd.name)))
			b.WriteString(descStyle.Render(cmd.desc))
		} else {
			b.WriteString(normalStyle.Render(fmt.Sprintf("    %-12s", cmd.name)))
			b.WriteString(descStyle.Render(cmd.desc))
		}
		b.WriteString("\n")
	}
	b.WriteString(grayStyle.Render("  ↑↓ navigate · Enter select"))
	b.WriteString("\n")
	return b.String()
}

func (m model) renderMenu() string {
	if len(m.menu.items) == 0 {
		return ""
	}

	var b strings.Builder

	titles := map[menuKind]string{menuNovel: "Novels", menuTheme: "Themes", menuChapter: "Chapters", menuResume: "Resume Reading"}
	accentColor := m.theme.AccentColor()
	titleStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a"))
	selectedStyle := lipgloss.NewStyle().Foreground(accentColor).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#d1d5db"))

	title := titles[m.menu.kind]
	b.WriteString(titleStyle.Render(fmt.Sprintf("  %s", title)))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render("  ┌" + strings.Repeat("─", 40) + "┐"))
	b.WriteString("\n")

	maxVisible := 8
	start := 0
	if m.menu.selected >= maxVisible {
		start = m.menu.selected - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.menu.items) {
		end = len(m.menu.items)
	}

	for i := start; i < end; i++ {
		item := m.menu.items[i]
		padLen := 39 - len(item.label)
		if padLen < 0 {
			padLen = 0
		}
		if i == m.menu.selected {
			b.WriteString(borderStyle.Render("  │") + selectedStyle.Render(fmt.Sprintf(" ● %s", item.label)) + strings.Repeat(" ", padLen) + borderStyle.Render("│"))
		} else {
			b.WriteString(borderStyle.Render("  │") + normalStyle.Render(fmt.Sprintf("   %s", item.label)) + strings.Repeat(" ", padLen) + borderStyle.Render("│"))
		}
		b.WriteString("\n")
	}

	b.WriteString(borderStyle.Render("  ╰" + strings.Repeat("─", 40) + "╯"))
	b.WriteString("\n")
	b.WriteString(grayStyle.Render("  ↑↓ navigate · Enter select · Esc cancel"))
	b.WriteString("\n")

	return b.String()
}

func renderSeparator(width int) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4b5563")).
		Render(strings.Repeat("─", width))
}
