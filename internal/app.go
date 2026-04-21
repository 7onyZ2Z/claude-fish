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

type appState int

const (
	stateReading appState = iota
	stateBoss
	stateMenu
)

type menuKind int

const (
	menuNovel menuKind = iota
	menuTheme
	menuChapter
)

type menuItem struct {
	label string
	value string
}

type menuState struct {
	active   bool
	kind     menuKind
	items    []menuItem
	selected int
}

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
	ti.Placeholder = "Type a message... or /help"
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

		// Menu navigation
		if m.state == stateMenu {
			return m.handleMenuKey(key)
		}

		// Direct shortcuts that don't conflict with typing
		switch key {
		case " ":
			if m.state == stateReading && m.pager != nil {
				m.pager.NextPage()
			}
			return m, nil
		case "tab":
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
		case "esc":
			if m.state == stateBoss {
				m.state = stateReading
				m.boss.Deactivate()
				return m, nil
			}
			return m, tea.Quit
		case "enter":
			return m.handleEnter()
		}

		// Everything else goes to text input
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
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

// ── Enter: process command or clear ──────────────────────────────

func (m model) handleEnter() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.input.Value())
	m.input.SetValue("")

	if !strings.HasPrefix(input, "/") {
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
		return m.openNovelMenu()
	case "/theme":
		return m.openThemeMenu()
	case "/chapters":
		if arg != "" {
			n, err := strconv.Atoi(arg)
			if err == nil && m.pager != nil {
				chapters := m.pager.Chapters()
				if n >= 1 && n <= len(chapters) {
					m.pager.GoToChapter(n - 1)
				}
			}
			return m, nil
		}
		return m.openChapterMenu()
	case "/back", "/b":
		if m.pager != nil {
			m.pager.PrevPage()
		}
		return m, nil
	case "/next", "/n":
		if m.pager != nil {
			m.pager.NextPage()
		}
		return m, nil
	case "/help":
		m.input.SetValue("/novel  /theme  /chapters [N]  /back  /next  /help")
		return m, nil
	}

	return m, nil
}

// ── Menu opening ─────────────────────────────────────────────────

func (m model) openNovelMenu() (tea.Model, tea.Cmd) {
	items := scanNovelDir(m.novelDir)
	if len(items) == 0 {
		m.input.SetValue("No novels found in ./" + m.novelDir + "/")
		return m, nil
	}
	m.state = stateMenu
	m.menu = menuState{active: true, kind: menuNovel, items: items, selected: 0}
	return m, nil
}

func (m model) openThemeMenu() (tea.Model, tea.Cmd) {
	var items []menuItem
	for _, t := range m.themes {
		items = append(items, menuItem{label: t.Name(), value: t.Name()})
	}
	m.state = stateMenu
	m.menu = menuState{active: true, kind: menuTheme, items: items, selected: m.themeIndex}
	return m, nil
}

func (m model) openChapterMenu() (tea.Model, tea.Cmd) {
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
	m.state = stateMenu
	m.menu = menuState{active: true, kind: menuChapter, items: items, selected: m.pager.Chapter()}
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
		m.state = stateReading
		m.menu = menuState{}
	}
	return m, nil
}

func (m model) confirmMenuSelection() (tea.Model, tea.Cmd) {
	if m.menu.selected < 0 || m.menu.selected >= len(m.menu.items) {
		m.state = stateReading
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
		}
	}

	m.state = stateReading
	m.menu = menuState{}
	return m, nil
}

// ── Helpers ──────────────────────────────────────────────────────

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

	switch m.state {
	case stateMenu:
		// Show content behind the menu (dimmed)
		content := m.renderReadingContent()
		menuView := m.renderMenu()
		return content + "\n" + menuView + "\n" + inputView + "\n" + renderSeparator(m.width)

	case stateReading:
		return m.renderReadingContent() + "\n" + renderSeparator(m.width) + "\n" + inputView + "\n" + renderSeparator(m.width)

	case stateBoss:
		s := m.boss.Streamer()
		visible := s.VisibleContent()
		highlighted := HighlightCode(visible, s.FileName())
		return m.theme.RenderCode(theme.CodeInfo{
			FileName:  s.FileName(),
			Content:   highlighted,
			Displayed: s.Displayed(),
			Total:     s.Total(),
			ThemeName: m.theme.Name(),
			Version:   m.version,
		}, m.width, m.height) + "\n" + renderSeparator(m.width) + "\n" + inputView + "\n" + renderSeparator(m.width)
	}

	return ""
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

func (m model) renderMenu() string {
	if len(m.menu.items) == 0 {
		return ""
	}

	var b strings.Builder

	// Menu title
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

	// Show at most 8 items, windowed around selected
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
		if i == m.menu.selected {
			b.WriteString(borderStyle.Render("  │") + selectedStyle.Render(fmt.Sprintf(" ● %s", item.label)) + strings.Repeat(" ", max(0, 39-len(item.label))) + borderStyle.Render("│"))
		} else {
			b.WriteString(borderStyle.Render("  │") + normalStyle.Render(fmt.Sprintf("   %s", item.label)) + strings.Repeat(" ", max(0, 39-len(item.label))) + borderStyle.Render("│"))
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
