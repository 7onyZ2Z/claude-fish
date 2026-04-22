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
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")

	var current *txtChapter

	for _, raw := range lines {
		line := strings.TrimSpace(raw)

		if isChapterHeading(line) {
			if current != nil {
				r.chapters = append(r.chapters, *current)
			}
			current = &txtChapter{title: line}
		} else if line != "" {
			if current == nil {
				current = &txtChapter{title: "序章"}
			}
			current.lines = append(current.lines, line)
		}
	}
	if current != nil {
		r.chapters = append(r.chapters, *current)
	}
}

func isChapterHeading(line string) bool {
	if len(line) == 0 || len(line) > 80 {
		return false
	}

	// Skip lines that look like body text (contain common punctuation mid-line)
	if strings.Contains(line, "，") || strings.Contains(line, "。") ||
		strings.Contains(line, "！") || strings.Contains(line, "？") ||
		strings.Contains(line, "、") || strings.Contains(line, "；") ||
		strings.Contains(line, "：") {
		return false
	}

	// 第X章/节/回/卷/部/篇 ...
	if isChineseChapterHeading(line) {
		return true
	}

	// Chapter/CHAPTER/chapter N
	lower := strings.ToLower(line)
	if strings.HasPrefix(lower, "chapter ") {
		return true
	}

	// 分节阅读 N
	if strings.HasPrefix(line, "分节阅读 ") {
		return true
	}

	return false
}

func isChineseChapterHeading(line string) bool {
	if !strings.HasPrefix(line, "第") {
		return false
	}
	rest := line[3:] // skip "第" (UTF-8: 3 bytes)

	// Consume Chinese or Arabic digits
	hasDigit := false
	for _, r := range rest {
		if isChineseDigit(r) || (r >= '0' && r <= '9') {
			hasDigit = true
			continue
		}
		break
	}
	if !hasDigit {
		return false
	}

	// Check for chapter-type suffix after digits
	// Find position after digits
	runes := []rune(rest)
	pos := 0
	for pos < len(runes) && (isChineseDigit(runes[pos]) || (runes[pos] >= '0' && runes[pos] <= '9')) {
		pos++
	}
	if pos >= len(runes) {
		return false
	}

	suffixes := []rune{'章', '节', '回', '卷', '部', '篇', '集'}
	for _, s := range suffixes {
		if runes[pos] == s {
			return true
		}
	}
	return false
}

func isChineseDigit(r rune) bool {
	return strings.ContainsRune("零一二三四五六七八九十百千万亿〇壹贰叁肆伍陆柒捌玖拾佰仟", r)
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
	// Estimate ~2 wrapped lines per source line for preallocation
	estimated := len(lines) * 2
	if estimated < 64 {
		estimated = 64
	}
	result := make([]string, 0, estimated)
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
	var b strings.Builder
	currentWidth := 0

	for _, r := range line {
		charWidth := runeWidth(r)
		if currentWidth+charWidth > width && b.Len() > 0 {
			result = append(result, b.String())
			b.Reset()
			b.WriteRune(r)
			currentWidth = charWidth
		} else {
			b.WriteRune(r)
			currentWidth += charWidth
		}
	}
	if b.Len() > 0 {
		result = append(result, b.String())
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
