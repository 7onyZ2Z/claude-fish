package internal

var defaultBossContent = []Segment{
	ThinkSegment(`I'll analyze the current codebase structure and implement the requested changes. Let me first understand the existing architecture by examining the key files.`),
	ThinkSegment(`Looking at the main entry point and the configuration module, I can see the application follows a clean architecture pattern with separate packages for internal logic, theme rendering, and file reading.`),
	ThinkSegment(`Now I'll implement the changes. I need to:
1. Add proper error handling in the reader module
2. Refactor the pagination logic to support variable-width characters
3. Update the theme interface to include status bar rendering
4. Add unit tests for the new functionality`),
	TextSegment(`I'll implement the changes now. Starting with the reader module:`),
	CodeSegment("internal/reader/reader.go", `package reader

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type Reader interface {
	Load(path string) error
	Chapters() []Chapter
	ReadPage(chapter, page, width, linesPerPage int) string
	TotalPages(chapter, width, linesPerPage int) int
}

type Chapter struct {
	Title string
	Index int
}

func WrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	var result []string
	var current string
	currentWidth := 0

	for _, r := range line {
		charWidth := RuneWidth(r)
		if currentWidth+charWidth > width && current != "" {
			result = append(result, current)
			current = string(r)
			currentWidth = charWidth
		} else {
			current += string(r)
			currentWidth += charWidth
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func RuneWidth(r rune) int {
	if r >= 0x1100 && (r <= 0x115F || r == 0x2329 || r == 0x232A ||
		(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE30 && r <= 0xFE6F) ||
		(r >= 0xFF01 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x20000 && r <= 0x2FFFD) ||
		(r >= 0x30000 && r <= 0x3FFFD)) {
		return 2
	}
	return 1
}

func WrapLines(lines []string, width int) []string {
	var result []string
	for _, line := range lines {
		result = append(result, WrapLine(line, width)...)
	}
	return result
}

func DetectEncoding(data []byte) string {
	if utf8.Valid(data) {
		return "utf-8"
	}
	return "gbk"
}

func FormatProgress(current, total int) string {
	if total <= 0 {
		return ""
	}
	pct := float64(current) / float64(total) * 100
	filled := int(pct / 5)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 20-filled)
	return fmt.Sprintf("[%s] %.0f%%", bar, pct)
}`),
	TextSegment("Now let me update the main application module with the new pagination logic:"),
	CodeSegment("internal/app.go", `func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if key == "tab" {
			return m.handleTab()
		}
		if key == "esc" {
			return m, tea.Quit
		}
		if key == "enter" {
			return m.handleEnter()
		}
		if key == " " && strings.TrimSpace(m.input.Value()) == "" {
			if m.pager != nil {
				m.pager.NextPage()
				m.saveHistory()
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}`),
	ThinkSegment(`The pagination logic looks correct. The CJK-aware wrapping ensures proper display of double-width characters. Now I need to verify the theme rendering doesn't clip content when the terminal is resized.`),
	TextSegment("The changes are complete. Let me create a summary of what was modified:"),
	CodeSegment("internal/pager.go", `func (p *Pager) NextPage() bool {
	if p.currentPg < p.totalPages[p.currentCh]-1 {
		p.currentPg++
		return true
	}
	if p.currentCh < len(p.chapters)-1 {
		p.currentCh++
		p.currentPg = 0
		return true
	}
	return false
}

func (p *Pager) PrevPage() bool {
	if p.currentPg > 0 {
		p.currentPg--
		return true
	}
	if p.currentCh > 0 {
		p.currentCh--
		p.currentPg = p.totalPages[p.currentCh] - 1
		if p.currentPg < 0 {
			p.currentPg = 0
		}
		return true
	}
	return false
}

func (p *Pager) GoToChapter(ch int) {
	if ch >= 0 && ch < len(p.chapters) {
		p.currentCh = ch
		p.currentPg = 0
	}
}`),
	ThinkSegment(`All changes have been applied successfully. The implementation includes proper CJK character handling, improved pagination, and the new progress bar formatter. The code follows the existing patterns in the codebase.`),
	TextSegment("Done. I've successfully implemented the changes across 3 files with proper error handling and test coverage."),
}

// Segment types for AI output simulation
type SegmentType int

const (
	SegmentThink SegmentType = iota
	SegmentText
	SegmentCode
)

type Segment struct {
	Type     SegmentType
	Content  string
	FileName string
}

func ThinkSegment(content string) Segment {
	return Segment{Type: SegmentThink, Content: content}
}

func TextSegment(content string) Segment {
	return Segment{Type: SegmentText, Content: content}
}

func CodeSegment(fileName, content string) Segment {
	return Segment{Type: SegmentCode, Content: content, FileName: fileName}
}
