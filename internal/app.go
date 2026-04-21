package internal

import (
	"strings"
	"time"

	"claude-fish/internal/reader"
	"claude-fish/internal/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appState int

const (
	stateReading appState = iota
	stateBoss
)

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
}

// newModel creates the internal model. Used by NewApp and tests.
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
	}
}

// NewApp creates the Bubble Tea program.
func NewApp(r reader.Reader, th theme.Theme, code, codeFile string, speed int, fileName string) *tea.Program {
	m := newModel(r, th, NewBossMode(code, codeFile, speed), 80, 24, speed)
	m.fileName = fileName
	return tea.NewProgram(m, tea.WithAltScreen())
}

func (model) Init() tea.Cmd { return nil }

type tickMsg time.Time

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.pager != nil {
			usable := m.theme.UsableHeight(msg.Height)
			if usable < 1 {
				usable = 1
			}
			m.pager.Resize(msg.Width-4, usable)
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg.String())

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

func (m model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateReading:
		switch key {
		case " ", "right", "l":
			if m.pager != nil {
				m.pager.NextPage()
			}
		case "b", "left", "h":
			if m.pager != nil {
				m.pager.PrevPage()
			}
		case "s":
			m.themeIndex = (m.themeIndex + 1) % len(m.themes)
			m.theme = m.themes[m.themeIndex]
			if m.pager != nil {
				usable := m.theme.UsableHeight(m.height)
				if usable < 1 {
					usable = 1
				}
				m.pager.SetThemeLines(usable)
			}
		case "tab":
			if m.boss.HasCode() {
				m.state = stateBoss
				m.boss.Activate()
				s := m.boss.Streamer()
				if !s.Done() {
					return m, tea.Tick(time.Duration(s.JitterSpeed())*time.Millisecond,
						func(t time.Time) tea.Msg { return tickMsg(t) })
				}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case stateBoss:
		switch key {
		case "tab", "esc":
			m.state = stateReading
			m.boss.Deactivate()
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateReading:
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
			totalChapters = len(m.pager.Chapters())
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
		}, m.width, m.height) + "\n" + renderSeparator(m.width) + "\n❯ \n" + renderSeparator(m.width)

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
		}, m.width, m.height) + "\n" + renderSeparator(m.width) + "\n❯ \n" + renderSeparator(m.width)
	}

	return ""
}

func renderSeparator(width int) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4b5563")).
		Render(strings.Repeat("─", width))
}
