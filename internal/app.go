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
	{"/help", "Show available commands"},
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
	messages   []string // command output lines shown between content and input
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
		boss = NewBossMode("", "", speed)
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
		version:    "v1.0.0",
		input:      ti,
		novelDir:   "novel",
	}
}

func NewApp(r reader.Reader, th theme.Theme, code, codeFile string, speed int, fileName string) *tea.Program {
	m := newModel(r, th, NewBossMode(code, codeFile, speed), 80, 24, speed)
	m.fileName = fileName
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
		// Esc: in boss mode go back, otherwise quit
		if key == "esc" {
			if m.state == stateBoss {
				m.state = stateReading
				m.boss.Deactivate()
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
			if !s.Done() {
				s.Advance(1)
				return m, tea.Tick(time.Duration(s.JitterSpeed())*time.Millisecond,
					func(t time.Time) tea.Msg { return tickMsg(t) })
			}
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
		s := m.boss.Streamer()
		if !s.Done() {
			return m, tea.Tick(time.Duration(s.JitterSpeed())*time.Millisecond,
				func(t time.Time) tea.Msg { return tickMsg(t) })
		}
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
		m.addMessage(input)
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
		m.addMessage("went back one page")
	case "/next":
		if m.pager != nil {
			m.pager.NextPage()
		}
		m.addMessage("went to next page")
	case "/help":
		var lines []string
		for _, c := range allCommands {
			lines = append(lines, fmt.Sprintf("  %-12s %s", c.name, c.desc))
		}
		m.addMessage(strings.Join(lines, "\n"))
	default:
		m.addMessage(fmt.Sprintf("unknown command: %s", cmd))
	}

	return m, nil
}

func (m *model) addMessage(msg string) {
	m.messages = append(m.messages, msg)
	// Keep only last 6 messages to avoid overflowing the screen
	if len(m.messages) > 6 {
		m.messages = m.messages[len(m.messages)-6:]
	}
}

// ── Command implementations ──────────────────────────────────────

func (m model) cmdNovel() (tea.Model, tea.Cmd) {
	items := scanNovelDir(m.novelDir)
	if len(items) == 0 {
		m.addMessage(fmt.Sprintf("No novels found in ./%s/", m.novelDir))
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
				m.addMessage(fmt.Sprintf("jumped to chapter %d: %s", n, chapters[n-1].Title))
			} else {
				m.addMessage(fmt.Sprintf("invalid chapter: %d (1-%d)", n, len(chapters)))
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
				m.addMessage(fmt.Sprintf("loaded: %s", filepath.Base(path)))
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
				m.addMessage(fmt.Sprintf("switched to theme: %s", t.Name()))
				break
			}
		}
	case menuChapter:
		ch, _ := strconv.Atoi(item.value)
		if m.pager != nil {
			title := m.pager.Chapters()[ch].Title
			m.pager.GoToChapter(ch)
			m.addMessage(fmt.Sprintf("jumped to: %s", title))
		}
	}

	m.menu = menuState{}
	return m, nil
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
		highlighted := HighlightCode(visible, s.FileName())
		content = m.theme.RenderCode(theme.CodeInfo{
			FileName:  s.FileName(),
			Content:   highlighted,
			Displayed: s.Displayed(),
			Total:     s.Total(),
			ThemeName: m.theme.Name(),
			Version:   m.version,
		}, m.width, m.height)
	}

	// 2. Message area (command output between content and input)
	msgView := m.renderMessages()

	// 3. Autocomplete popup or submenu
	popupView := m.renderPopup()

	return content + msgView + popupView + renderSeparator(m.width) + "\n" + inputView + "\n" + renderSeparator(m.width)
}

func (m model) renderReadingContent() string {
	title := ""
	content := ""
	pageNum := 0
	totalPages := 0
	totalChapters := 0
	if m.pager != nil {
		title = m.pager.CurrentTitle()
		content = m.pager.CurrentContent()
		pageNum = m.pager.Page()
		totalPages = m.pager.TotalPages()
		totalChapters = m.pager.TotalChapters()
	}
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

func (m model) renderMessages() string {
	if len(m.messages) == 0 {
		return ""
	}
	grayStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
	var b strings.Builder
	for _, msg := range m.messages {
		b.WriteString(grayStyle.Render("  " + msg))
		b.WriteString("\n")
	}
	return b.String()
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

	titles := map[menuKind]string{menuNovel: "Novels", menuTheme: "Themes", menuChapter: "Chapters"}
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
