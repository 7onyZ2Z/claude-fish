package reader

import (
	"os"
	"strings"
)

type MarkdownReader struct {
	chapters []txtChapter
}

func (r *MarkdownReader) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	r.parse(string(data))
	return nil
}

func (r *MarkdownReader) parse(content string) {
	r.chapters = nil
	lines := strings.Split(content, "\n")
	var current *txtChapter

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "# ") || strings.HasPrefix(trimmed, "## ") {
			if current != nil {
				r.chapters = append(r.chapters, *current)
			}
			title := strings.TrimLeft(trimmed, "# ")
			title = strings.TrimSpace(title)
			current = &txtChapter{title: title}
		} else if trimmed != "" {
			if current == nil {
				current = &txtChapter{title: "序章"}
			}
			current.lines = append(current.lines, trimmed)
		}
	}
	if current != nil {
		r.chapters = append(r.chapters, *current)
	}
}

func (r *MarkdownReader) Chapters() []Chapter {
	result := make([]Chapter, len(r.chapters))
	for i, c := range r.chapters {
		result[i] = Chapter{Title: c.title, Index: i}
	}
	return result
}

func (r *MarkdownReader) ReadPage(chapter, page, width, linesPerPage int) string {
	if chapter < 0 || chapter >= len(r.chapters) {
		return ""
	}
	lines := wrapLines(r.chapters[chapter].lines, width)
	start := page * linesPerPage
	end := start + linesPerPage
	if end > len(lines) {
		end = len(lines)
	}
	if start >= len(lines) {
		return ""
	}
	return strings.Join(lines[start:end], "\n")
}

func (r *MarkdownReader) TotalPages(chapter, width, linesPerPage int) int {
	if chapter < 0 || chapter >= len(r.chapters) {
		return 0
	}
	lines := wrapLines(r.chapters[chapter].lines, width)
	return (len(lines) + linesPerPage - 1) / linesPerPage
}
