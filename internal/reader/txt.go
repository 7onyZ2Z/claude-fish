package reader

import (
	"bytes"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type TXTReader struct {
	chapters []txtChapter
}

type txtChapter struct {
	title string
	lines []string
}

func (r *TXTReader) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := decodeBytes(data)
	r.parse(content)
	return nil
}

func decodeBytes(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	rd := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(rd)
	if err != nil {
		return string(data)
	}
	return string(decoded)
}

func (r *TXTReader) parse(content string) {
	r.chapters = nil
	paragraphs := strings.Split(content, "\n\n")
	var current *txtChapter

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		lines := strings.Split(p, "\n")
		first := strings.TrimSpace(lines[0])

		if isChapterHeading(first) {
			if current != nil {
				r.chapters = append(r.chapters, *current)
			}
			current = &txtChapter{title: first}
			for _, l := range lines[1:] {
				if l = strings.TrimSpace(l); l != "" {
					current.lines = append(current.lines, l)
				}
			}
		} else {
			if current == nil {
				current = &txtChapter{title: "序章"}
			}
			for _, l := range lines {
				if l = strings.TrimSpace(l); l != "" {
					current.lines = append(current.lines, l)
				}
			}
		}
	}
	if current != nil {
		r.chapters = append(r.chapters, *current)
	}
}

func isChapterHeading(line string) bool {
	prefixes := []string{"第", "Chapter", "chapter", "CHAPTER"}
	for _, p := range prefixes {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	return false
}

func (r *TXTReader) Chapters() []Chapter {
	result := make([]Chapter, len(r.chapters))
	for i, c := range r.chapters {
		result[i] = Chapter{Title: c.title, Index: i}
	}
	return result
}

func (r *TXTReader) ReadPage(chapter, page, width, linesPerPage int) string {
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

func (r *TXTReader) TotalPages(chapter, width, linesPerPage int) int {
	if chapter < 0 || chapter >= len(r.chapters) {
		return 0
	}
	lines := wrapLines(r.chapters[chapter].lines, width)
	return (len(lines) + linesPerPage - 1) / linesPerPage
}

func wrapLines(lines []string, width int) []string {
	var result []string
	for _, line := range lines {
		result = append(result, wrapLine(line, width)...)
	}
	return result
}

func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}
	var result []string
	var current string
	currentWidth := 0

	for _, r := range line {
		charWidth := runeWidth(r)
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

func runeWidth(r rune) int {
	if r >= 0x1100 && (r <= 0x115F || r == 0x2329 || r == 0x232A ||
		(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
		(r >= 0xAC00 && r <= 0xD7A3) ||
		(r >= 0xF900 && r <= 0xFAFF) ||
		(r >= 0xFE10 && r <= 0xFE19) ||
		(r >= 0xFE30 && r <= 0xFE6F) ||
		(r >= 0xFF01 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) ||
		(r >= 0x20000 && r <= 0x2FFFD) ||
		(r >= 0x30000 && r <= 0x3FFFD)) {
		return 2
	}
	return 1
}
