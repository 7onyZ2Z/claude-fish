# claude-fish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a terminal novel reader that mimics CLI programming tools, with a boss key that instantly switches to a realistic code streaming view.

**Architecture:** Go Bubble Tea TUI app with a state machine (Welcome → Reading → BossMode). Theme interface enables swappable CLI visual styles. Reader interface supports EPUB/TXT/Markdown. Pager handles CJK-aware pagination. Streamer simulates realistic code output with syntax highlighting.

**Tech Stack:** Go, Bubble Tea, Lip Gloss, Cobra, Chroma, stdlib archive/zip + encoding/xml for EPUB

---

## File Structure

| File | Responsibility |
|------|---------------|
| `main.go` | Entry point, calls cmd.Execute() |
| `cmd/root.go` | Cobra command, flag parsing |
| `internal/theme/theme.go` | Theme interface + data types |
| `internal/theme/claudecode.go` | Claude Code visual style |
| `internal/theme/codex.go` | Codex CLI visual style |
| `internal/theme/opencode.go` | opencode visual style |
| `internal/reader/reader.go` | Reader interface + Chapter type |
| `internal/reader/txt.go` | TXT format parser |
| `internal/reader/markdown.go` | Markdown format parser |
| `internal/reader/epub.go` | EPUB format parser |
| `internal/pager.go` | CJK-aware pagination engine |
| `internal/streamer.go` | Code streaming logic (pure functions) |
| `internal/boss.go` | Boss mode state |
| `internal/app.go` | Bubble Tea model, state machine, key handling |
| `testdata/sample.txt` | Test fixture |
| `testdata/sample.md` | Test fixture |

---

### Task 1: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `.gitignore`
- Create: `testdata/sample.txt`
- Create: `testdata/sample.md`

- [ ] **Step 1: Initialize Go module and create directories**

```bash
cd /Users/tony/Code/claude-fish
go mod init claude-fish
mkdir -p internal/theme internal/reader cmd testdata
```

- [ ] **Step 2: Add dependencies**

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
go get github.com/spf13/cobra
go get github.com/alecthomas/chroma
go get github.com/mattn/go-runewidth
```

- [ ] **Step 3: Create .gitignore**

```
# .gitignore
claude-fish
*.exe
.superpowers/
```

- [ ] **Step 4: Create test fixtures**

`testdata/sample.txt`:
```
第一章 开始

这是第一章的内容。
讲述了一个关于编程的故事。
主角是一个年轻的程序员。

第二章 发展

这是第二章的内容。
故事继续发展。
代码开始出现在生活中。
```

`testdata/sample.md`:
```
# 第一章 开始

这是第一章的内容。讲述了一个关于编程的故事。

# 第二章 发展

这是第二章的内容。故事继续发展。
```

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "chore: initialize project with go module and dependencies"
```

---

### Task 2: Theme Interface & Types

**Files:**
- Create: `internal/theme/theme.go`

- [ ] **Step 1: Write theme interface and data types**

```go
// internal/theme/theme.go
package theme

import "github.com/charmbracelet/lipgloss"

type WelcomeInfo struct {
	Version   string
	FileName  string
	Chapters  int
	ThemeName string
}

type PageInfo struct {
	ChapterTitle string
	Content      string
	PageNum      int
	TotalPages   int
	FileName     string
}

type CodeInfo struct {
	FileName  string
	Content   string
	Displayed int
	Total     int
}

type KeyHint struct {
	Key  string
	Desc string
}

type StatusInfo struct {
	Mode  string // "welcome", "reading", "boss"
	Hints []KeyHint
}

type Theme interface {
	Name() string
	RenderWelcome(info WelcomeInfo) string
	RenderPage(info PageInfo, width, height int) string
	RenderCode(info CodeInfo, width, height int) string
	RenderStatusBar(info StatusInfo, width int) string
	AccentColor() lipgloss.Color
	UsableHeight(termHeight int) int
}

// All returns all available themes in order.
func All() []Theme {
	return []Theme{
		NewClaudeCode(),
		NewCodex(),
		NewOpenCode(),
	}
}

// FindByName returns the theme with the given name, or the first theme if not found.
func FindByName(name string) Theme {
	for _, t := range All() {
		if t.Name() == name {
			return t
		}
	}
	return All()[0]
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/theme/
```

Expected: compile error about undefined `NewClaudeCode`, `NewCodex`, `NewOpenCode` — this is expected, we'll implement them later.

- [ ] **Step 3: Create stub implementations so it compiles**

Create `internal/theme/stub.go` (temporary, removed when real themes are implemented):

```go
// internal/theme/stub.go
// Temporary stubs — replaced by real theme implementations in Tasks 7, 11, 12.
package theme

import "github.com/charmbracelet/lipgloss"

type stubTheme struct{}

func (stubTheme) Name() string                                    { return "stub" }
func (stubTheme) RenderWelcome(WelcomeInfo) string                { return "" }
func (stubTheme) RenderPage(PageInfo, int, int) string            { return "" }
func (stubTheme) RenderCode(CodeInfo, int, int) string            { return "" }
func (stubTheme) RenderStatusBar(StatusInfo, int) string          { return "" }
func (stubTheme) AccentColor() lipgloss.Color                     { return "#7c3aed" }
func (stubTheme) UsableHeight(termHeight int) int                 { return termHeight - 8 }

func NewClaudeCode() Theme { return stubTheme{} }
func NewCodex() Theme      { return stubTheme{} }
func NewOpenCode() Theme   { return stubTheme{} }
```

- [ ] **Step 4: Verify it compiles**

```bash
go build ./...
```

Expected: SUCCESS

- [ ] **Step 5: Commit**

```bash
git add internal/theme/
git commit -m "feat: add Theme interface and data types"
```

---

### Task 3: TXT Reader

**Files:**
- Create: `internal/reader/reader.go`
- Create: `internal/reader/txt.go`
- Create: `internal/reader/txt_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/reader/txt_test.go
package reader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTXTReaderLoadAndChapters(t *testing.T) {
	r := &TXTReader{}
	path := filepath.Join("..", "..", "testdata", "sample.txt")
	if err := r.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	chapters := r.Chapters()
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}
	if chapters[0].Title != "第一章 开始" {
		t.Errorf("chapter 0 title = %q, want %q", chapters[0].Title, "第一章 开始")
	}
	if chapters[1].Title != "第二章 发展" {
		t.Errorf("chapter 1 title = %q, want %q", chapters[1].Title, "第二章 发展")
	}
}

func TestTXTReaderPages(t *testing.T) {
	r := &TXTReader{}
	path := filepath.Join("..", "..", "testdata", "sample.txt")
	if err := r.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// With a small line limit (5 lines per page), chapter 0 should have multiple pages
	totalPages := r.TotalPages(0, 80, 5)
	if totalPages < 1 {
		t.Errorf("TotalPages(0) = %d, want >= 1", totalPages)
	}

	page0 := r.ReadPage(0, 0, 80, 5)
	if page0 == "" {
		t.Error("ReadPage(0, 0) returned empty string")
	}
}

func TestTXTReaderInvalidPath(t *testing.T) {
	r := &TXTReader{}
	err := r.Load("nonexistent.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/reader/ -v -run TestTXTReader
```

Expected: FAIL — `TXTReader` type not defined

- [ ] **Step 3: Write Reader interface**

```go
// internal/reader/reader.go
package reader

// Chapter represents a single chapter in a novel.
type Chapter struct {
	Title string
	Index int
}

// Reader is the interface for novel format parsers.
// Page dimensions (width, linesPerPage) are passed to ReadPage/TotalPages
// so implementations can do their own line-wrapping.
type Reader interface {
	Load(path string) error
	Chapters() []Chapter
	ReadPage(chapter, page, width, linesPerPage int) string
	TotalPages(chapter, width, linesPerPage int) int
}
```

- [ ] **Step 4: Write TXT implementation**

```go
// internal/reader/txt.go
package reader

import (
	"os"
	"strings"
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
	r.parse(string(data))
	return nil
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

		// Heuristic: if the first line looks like a chapter heading
		if isChapterHeading(first) {
			if current != nil {
				r.chapters = append(r.chapters, *current)
			}
			current = &txtChapter{title: first}
			// Remaining lines in this paragraph are content
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

// wrapLines wraps lines to fit within width, accounting for CJK double-width characters.
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
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/reader/ -v -run TestTXTReader
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/reader/
git commit -m "feat: add Reader interface and TXT parser with CJK-aware wrapping"
```

---

### Task 4: Markdown Reader

**Files:**
- Create: `internal/reader/markdown.go`
- Create: `internal/reader/markdown_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/reader/markdown_test.go
package reader

import (
	"path/filepath"
	"testing"
)

func TestMarkdownReaderLoadAndChapters(t *testing.T) {
	r := &MarkdownReader{}
	path := filepath.Join("..", "..", "testdata", "sample.md")
	if err := r.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	chapters := r.Chapters()
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}
	if chapters[0].Title != "第一章 开始" {
		t.Errorf("chapter 0 title = %q, want %q", chapters[0].Title, "第一章 开始")
	}
}

func TestMarkdownReaderPages(t *testing.T) {
	r := &MarkdownReader{}
	path := filepath.Join("..", "..", "testdata", "sample.md")
	if err := r.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	totalPages := r.TotalPages(0, 80, 5)
	if totalPages < 1 {
		t.Errorf("TotalPages(0) = %d, want >= 1", totalPages)
	}

	page0 := r.ReadPage(0, 0, 80, 5)
	if page0 == "" {
		t.Error("ReadPage(0, 0) returned empty string")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/reader/ -v -run TestMarkdownReader
```

Expected: FAIL — `MarkdownReader` type not defined

- [ ] **Step 3: Write Markdown implementation**

```go
// internal/reader/markdown.go
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
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/reader/ -v -run TestMarkdownReader
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/reader/markdown.go internal/reader/markdown_test.go
git commit -m "feat: add Markdown parser"
```

---

### Task 5: EPUB Reader

**Files:**
- Create: `internal/reader/epub.go`
- Create: `internal/reader/epub_test.go`

Note: Uses stdlib `archive/zip` and `encoding/xml` — no third-party EPUB library needed.

- [ ] **Step 1: Create a minimal EPUB test fixture**

Create `testdata/sample.epub` — this is a ZIP file with the minimal EPUB structure. We'll generate it in the test setup.

Actually, we create it programmatically in the test. No fixture file needed — the test builds a minimal EPUB in a temp dir.

- [ ] **Step 2: Write the failing test**

```go
// internal/reader/epub_test.go
package reader

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestEPUB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	epubPath := filepath.Join(dir, "test.epub")

	f, err := os.Create(epubPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)

	// mimetype (must be first, uncompressed)
	mw, _ := w.Create("mimetype")
	mw.Write([]byte("application/epub+zip"))

	// META-INF/container.xml
	cw, _ := w.Create("META-INF/container.xml")
	cw.Write([]byte(`<?xml version="1.0"?>
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container" version="1.0">
  <rootfiles>
    <rootfile full-path="content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	// content.opf
	ow, _ := w.Create("content.opf")
	ow.Write([]byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
  </metadata>
  <manifest>
    <item id="ch1" href="ch1.xhtml" media-type="application/xhtml+xml"/>
    <item id="ch2" href="ch2.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
    <itemref idref="ch2"/>
  </spine>
</package>`))

	// ch1.xhtml
	c1w, _ := w.Create("ch1.xhtml")
	c1w.Write([]byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter 1</title></head>
<body><p>这是第一章的内容。讲述了一个测试故事。</p><p>第二段内容在这里。</p></body>
</html>`))

	// ch2.xhtml
	c2w, _ := w.Create("ch2.xhtml")
	c2w.Write([]byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter 2</title></head>
<body><p>这是第二章的内容。故事继续发展。</p></body>
</html>`))

	w.Close()
	return epubPath
}

func TestEPUBReaderLoadAndChapters(t *testing.T) {
	r := &EPUBReader{}
	path := createTestEPUB(t)
	if err := r.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	chapters := r.Chapters()
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}
	if !strings.Contains(chapters[0].Title, "1") {
		t.Errorf("chapter 0 title = %q, want to contain '1'", chapters[0].Title)
	}
}

func TestEPUBReaderPages(t *testing.T) {
	r := &EPUBReader{}
	path := createTestEPUB(t)
	if err := r.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	totalPages := r.TotalPages(0, 80, 5)
	if totalPages < 1 {
		t.Errorf("TotalPages(0) = %d, want >= 1", totalPages)
	}

	page0 := r.ReadPage(0, 0, 80, 5)
	if page0 == "" {
		t.Error("ReadPage(0, 0) returned empty string")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/reader/ -v -run TestEPUBReader
```

Expected: FAIL — `EPUBReader` type not defined

- [ ] **Step 4: Write EPUB implementation**

```go
// internal/reader/epub.go
package reader

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"os"
	"regexp"
	"strings"
)

type EPUBReader struct {
	chapters []txtChapter
}

// EPUB XML structures
type container struct {
	XMLName   xml.Name       `xml:"container"`
	RootFiles []containerRoot `xml:"rootfiles>rootfile"`
}

type containerRoot struct {
	FullPath string `xml:"full-path,attr"`
}

type opfPackage struct {
	XMLName xml.Name  `xml:"package"`
	Manifest []opfItem `xml:"manifest>item"`
	Spine   []opfItemRef `xml:"spine>itemref"`
}

type opfItem struct {
	ID       string `xml:"id,attr"`
	Href     string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

type opfItemRef struct {
	IDRef string `xml:"idref,attr"`
}

func (r *EPUBReader) Load(path string) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zr.Close()

	// 1. Find OPF from container.xml
	opfPath := r.findOPF(&zr.Reader)

	// 2. Parse OPF to get spine order
	pkg := r.parseOPF(&zr.Reader, opfPath)

	// 3. Build manifest map
	manifest := make(map[string]string) // id -> href
	for _, item := range pkg.Manifest {
		if item.MediaType == "application/xhtml+xml" {
			manifest[item.ID] = item.Href
		}
	}

	// 4. Follow spine, extract text from each chapter
	r.chapters = nil
	for i, ref := range pkg.Spine {
		href, ok := manifest[ref.IDRef]
		if !ok {
			continue
		}
		content := r.extractText(&zr.Reader, href)
		lines := strings.Split(content, "\n")
		var nonEmpty []string
		for _, l := range lines {
			if l = strings.TrimSpace(l); l != "" {
				nonEmpty = append(nonEmpty, l)
			}
		}
		title := extractTitle(content)
		if title == "" {
			title = "Chapter"
		}
		r.chapters = append(r.chapters, txtChapter{
			title: title,
			lines: nonEmpty,
		})
		if i == 0 && len(r.chapters) > 0 {
			// If only one chapter with default title, number it
		}
	}

	// Number chapters
	for i := range r.chapters {
		if r.chapters[i].title == "Chapter" {
			// Keep as-is for now
		}
	}

	return nil
}

func (r *EPUBReader) findOPF(zr *zip.Reader) string {
	for _, f := range zr.File {
		if f.Name == "META-INF/container.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			var c container
			xml.Unmarshal(data, &c)
			if len(c.RootFiles) > 0 {
				return c.RootFiles[0].FullPath
			}
		}
	}
	return "content.opf" // fallback
}

func (r *EPUBReader) parseOPF(zr *zip.Reader, opfPath string) *opfPackage {
	for _, f := range zr.File {
		if f.Name == opfPath {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			var pkg opfPackage
			xml.Unmarshal(data, &pkg)
			return &pkg
		}
	}
	return &opfPackage{}
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)
var multiSpaceRe = regexp.MustCompile(`\s+`)

func (r *EPUBReader) extractText(zr *zip.Reader, href string) string {
	for _, f := range zr.File {
		if f.Name == href || strings.HasSuffix(f.Name, "/"+href) {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			text := string(data)
			// Replace block tags with newlines
			text = regexp.MustCompile(`<(?:p|br|div|h[1-6])[^>]*>`).ReplaceAllString(text, "\n")
			text = htmlTagRe.ReplaceAllString(text, "")
			text = multiSpaceRe.ReplaceAllString(text, " ")
			return strings.TrimSpace(text)
		}
	}
	return ""
}

var titleRe = regexp.MustCompile(`(?i)<title>([^<]+)</title>`)

func extractTitle(content string) string {
	matches := titleRe.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (r *EPUBReader) Chapters() []Chapter {
	result := make([]Chapter, len(r.chapters))
	for i, c := range r.chapters {
		result[i] = Chapter{Title: c.title, Index: i}
	}
	return result
}

func (r *EPUBReader) ReadPage(chapter, page, width, linesPerPage int) string {
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

func (r *EPUBReader) TotalPages(chapter, width, linesPerPage int) int {
	if chapter < 0 || chapter >= len(r.chapters) {
		return 0
	}
	lines := wrapLines(r.chapters[chapter].lines, width)
	return (len(lines) + linesPerPage - 1) / linesPerPage
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/reader/ -v -run TestEPUBReader
```

Expected: PASS

- [ ] **Step 6: Run all reader tests together**

```bash
go test ./internal/reader/ -v
```

Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add internal/reader/epub.go internal/reader/epub_test.go
git commit -m "feat: add EPUB parser using stdlib zip+xml"
```

---

### Task 6: Pager

**Files:**
- Create: `internal/pager.go`
- Create: `internal/pager_test.go`

The Pager wraps a Reader and provides convenient page navigation. It delegates line-wrapping to the Reader but manages chapter/page state and re-pagination on resize.

- [ ] **Step 1: Write the failing test**

```go
// internal/pager_test.go
package internal

import (
	"path/filepath"
	"testing"

	"claude-fish/internal/reader"
)

func TestPagerNavigation(t *testing.T) {
	r := &reader.TXTReader{}
	path := filepath.Join("..", "testdata", "sample.txt")
	if err := r.Load(path); err != nil {
		t.Fatal(err)
	}

	p := NewPager(r, 80, 5)

	chapters := p.Chapters()
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}

	// Should start at chapter 0, page 0
	if p.Chapter() != 0 || p.Page() != 0 {
		t.Errorf("initial state = ch%d pg%d, want ch0 pg0", p.Chapter(), p.Page())
	}

	// NextPage within chapter
	totalPages := p.TotalPages()
	for i := 0; i < totalPages; i++ {
		content := p.CurrentContent()
		if content == "" && i == 0 {
			t.Error("page 0 content is empty")
		}
		if i < totalPages-1 {
			if !p.NextPage() {
				t.Errorf("NextPage() returned false at page %d of %d", i, totalPages)
			}
		}
	}

	// PrevPage should work
	if totalPages > 1 {
		if !p.PrevPage() {
			t.Error("PrevPage() returned false")
		}
		if p.Page() != totalPages-2 {
			t.Errorf("after PrevPage, page = %d, want %d", p.Page(), totalPages-2)
		}
	}
}

func TestPagerResize(t *testing.T) {
	r := &reader.TXTReader{}
	path := filepath.Join("..", "testdata", "sample.txt")
	if err := r.Load(path); err != nil {
		t.Fatal(err)
	}

	p := NewPager(r, 80, 5)
	pagesBefore := p.TotalPages()

	p.Resize(40, 5) // Narrower width means more wrapped lines
	pagesAfter := p.TotalPages()

	if pagesAfter < pagesBefore {
		t.Errorf("narrower width should have >= pages, got %d < %d", pagesAfter, pagesBefore)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ -v -run TestPager
```

Expected: FAIL — `NewPager` not defined

- [ ] **Step 3: Write Pager implementation**

```go
// internal/pager.go
package internal

import (
	"claude-fish/internal/reader"
)

// Pager manages chapter/page navigation over a Reader.
type Pager struct {
	r             reader.Reader
	width         int
	linesPerPage  int
	currentCh     int
	currentPg     int
	chapters      []reader.Chapter
	totalPages    []int // cached totalPages per chapter
}

func NewPager(r reader.Reader, width, linesPerPage int) *Pager {
	chapters := r.Chapters()
	p := &Pager{
		r:            r,
		width:        width,
		linesPerPage: linesPerPage,
		currentCh:    0,
		currentPg:    0,
		chapters:     chapters,
		totalPages:   make([]int, len(chapters)),
	}
	p.recalcAll()
	return p
}

func (p *Pager) recalcAll() {
	for i := range p.chapters {
		p.totalPages[i] = p.r.TotalPages(i, p.width, p.linesPerPage)
	}
	// Clamp current page
	if p.currentPg >= p.totalPages[p.currentCh] && p.totalPages[p.currentCh] > 0 {
		p.currentPg = p.totalPages[p.currentCh] - 1
	}
}

func (p *Pager) Chapters() []reader.Chapter { return p.chapters }
func (p *Pager) Chapter() int               { return p.currentCh }
func (p *Pager) Page() int                  { return p.currentPg }

func (p *Pager) TotalPages() int {
	if p.currentCh < 0 || p.currentCh >= len(p.totalPages) {
		return 0
	}
	return p.totalPages[p.currentCh]
}

func (p *Pager) CurrentContent() string {
	return p.r.ReadPage(p.currentCh, p.currentPg, p.width, p.linesPerPage)
}

func (p *Pager) CurrentTitle() string {
	if p.currentCh < 0 || p.currentCh >= len(p.chapters) {
		return ""
	}
	return p.chapters[p.currentCh].Title
}

// NextPage advances to the next page. Returns false if at the last page of the last chapter.
func (p *Pager) NextPage() bool {
	if p.currentPg < p.totalPages[p.currentCh]-1 {
		p.currentPg++
		return true
	}
	// Try next chapter
	if p.currentCh < len(p.chapters)-1 {
		p.currentCh++
		p.currentPg = 0
		return true
	}
	return false
}

// PrevPage goes back one page. Returns false if at the first page of the first chapter.
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

func (p *Pager) Resize(width, linesPerPage int) {
	p.width = width
	p.linesPerPage = linesPerPage
	p.recalcAll()
}

func (p *Pager) SetThemeLines(linesPerPage int) {
	p.linesPerPage = linesPerPage
	p.recalcAll()
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/ -v -run TestPager
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pager.go internal/pager_test.go
git commit -m "feat: add Pager with chapter/page navigation and resize support"
```

---

### Task 7: Claude Code Theme

**Files:**
- Modify: `internal/theme/theme.go` (no changes needed, just verify interface match)
- Modify: `internal/theme/stub.go` → Replace with `internal/theme/claudecode.go`
- Create: `internal/theme/claudecode_test.go`
- Delete: `internal/theme/stub.go`

This is the most substantial task — the Claude Code theme renders the full visual layout.

- [ ] **Step 1: Write the failing test**

```go
// internal/theme/claudecode_test.go
package theme

import (
	"strings"
	"testing"
)

func TestClaudeCodeName(t *testing.T) {
	th := NewClaudeCode()
	if th.Name() != "claude" {
		t.Errorf("Name() = %q, want %q", th.Name(), "claude")
	}
}

func TestClaudeCodeRenderPage(t *testing.T) {
	th := NewClaudeCode()
	info := PageInfo{
		ChapterTitle: "第一章 开始",
		Content:      "这是测试内容。",
		PageNum:      1,
		TotalPages:   10,
		FileName:     "test.txt",
	}
	output := th.RenderPage(info, 80, 24)

	if !strings.Contains(output, "第一章 开始") {
		t.Error("RenderPage output missing chapter title")
	}
	if !strings.Contains(output, "这是测试内容") {
		t.Error("RenderPage output missing content")
	}
	if !strings.Contains(output, "1/10") {
		t.Error("RenderPage output missing page indicator")
	}
}

func TestClaudeCodeRenderCode(t *testing.T) {
	th := NewClaudeCode()
	info := CodeInfo{
		FileName:  "main.go",
		Content:   "package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n",
		Displayed: 30,
		Total:     60,
	}
	output := th.RenderCode(info, 80, 24)

	if !strings.Contains(output, "main.go") {
		t.Error("RenderCode output missing filename")
	}
}

func TestClaudeCodeRenderWelcome(t *testing.T) {
	th := NewClaudeCode()
	info := WelcomeInfo{
		Version:   "v1.0.0",
		FileName:  "三体.txt",
		Chapters:  42,
		ThemeName: "claude",
	}
	output := th.RenderWelcome(info)

	if !strings.Contains(output, "三体.txt") {
		t.Error("RenderWelcome output missing filename")
	}
}

func TestClaudeCodeUsableHeight(t *testing.T) {
	th := NewClaudeCode()
	usable := th.UsableHeight(24)
	if usable <= 0 || usable >= 24 {
		t.Errorf("UsableHeight(24) = %d, want between 1 and 23", usable)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/theme/ -v -run TestClaudeCode
```

Expected: FAIL — stub doesn't produce correct output

- [ ] **Step 3: Write the Claude Code theme implementation**

```go
// internal/theme/claudecode.go
package theme

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	claudePurple    = "#7c3aed"
	claudeLightPurple = "#a78bfa"
	claudeOrange    = "#f97316"
	claudeGreen     = "#22c55e"
	claudeDarkPurple = "#2d1b69"
	claudeDarkBg    = "#1a1a2e"
	claudeGray      = "#6b7280"
	claudeLightGray = "#d1d5db"
)

type claudeCodeTheme struct{}

func NewClaudeCode() Theme {
	return claudeCodeTheme{}
}

func (claudeCodeTheme) Name() string { return "claude" }

func (claudeCodeTheme) AccentColor() lipgloss.Color { return claudePurple }

func (claudeCodeTheme) UsableHeight(termHeight int) int {
	// header: 3 lines, status bar: 2 lines, separator: 2 lines, padding: 2
	return termHeight - 9
}

func (claudeCodeTheme) RenderWelcome(info WelcomeInfo) string {
	var b strings.Builder
	w := 80

	// Header box
	headerBorder := lipgloss.NewStyle().Foreground(lipgloss.Color(claudePurple))
	b.WriteString(headerBorder.Render(fmt.Sprintf("╭─── claude-fish %s %s╮", info.Version, strings.Repeat("─", w-22-len(info.Version)))))
	b.WriteString("\n")

	// Left side: welcome + logo
	logoLines := []string{
		"                      ▐▛███▜▌",
		"                     ▝▜█████▛▘",
		"                       ▘▘ ▝▝",
	}

	orangeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeOrange))
	accentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple))

	// Tips column
	tips := []string{
		fmt.Sprintf("Loaded %s", info.FileName),
		fmt.Sprintf("%d chapters found", info.Chapters),
		"Press Space to start reading",
		"Press Tab for boss mode",
	}

	leftLines := []string{
		"                  Welcome!",
		"",
	}
	leftLines = append(leftLines, logoLines...)
	leftLines = append(leftLines, "",
		fmt.Sprintf("   %s · %s", accentStyle.Render(info.ThemeName), info.Version),
	)

	// Merge left and right
	maxLines := len(leftLines)
	if len(tips) > maxLines {
		maxLines = len(tips)
	}

	for i := 0; i < maxLines; i++ {
		var left, right string
		if i < len(leftLines) {
			left = leftLines[i]
		}
		if i < len(tips) {
			right = tips[i]
		}

		leftStyled := left
		if strings.Contains(left, "▐") || strings.Contains(left, "▜") || strings.Contains(left, "▝") {
			leftStyled = orangeStyle.Render(left)
		}

		// Pad to make two columns
		leftPadded := fmt.Sprintf("%-44s", leftStyled)
		rightStyled := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGray)).Render(right)
		b.WriteString(headerBorder.Render("│") + leftPadded + "│ " + rightStyled + "\n")
	}

	b.WriteString(headerBorder.Render(fmt.Sprintf("╰%s╯", strings.Repeat("─", w-2))))
	b.WriteString("\n")
	return b.String()
}

func (claudeCodeTheme) RenderPage(info PageInfo, width, height int) string {
	var b strings.Builder

	// Header bar
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudePurple)).
		Border(lipgloss.RoundedBorder(), true, false, true, false).
		BorderForeground(lipgloss.Color(claudePurple)).
		Width(width - 2)

	headerText := fmt.Sprintf(" ● Reading %s │ %s │ Page %d/%d",
		info.FileName, info.ChapterTitle, info.PageNum+1, info.TotalPages)
	b.WriteString(headerStyle.Render(headerText))
	b.WriteString("\n")

	// Chapter title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudeLightPurple)).
		Bold(true)
	b.WriteString("✦ ")
	b.WriteString(titleStyle.Render(info.ChapterTitle))
	b.WriteString("\n")

	// Content in chat bubble
	bubbleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudeLightGray)).
		Background(lipgloss.Color(claudeDarkPurple)).
		BorderLeft(true).
		BorderBackground(lipgloss.Color(claudeDarkPurple)).
		BorderForeground(lipgloss.Color(claudePurple)).
		Width(width - 4)

	b.WriteString(bubbleStyle.Render(info.Content))
	b.WriteString("\n")

	// Tool indicator
	toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGreen))
	b.WriteString(toolStyle.Render("✔"))
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGray)).Render(
		fmt.Sprintf(" Page %d of %d", info.PageNum+1, info.TotalPages)))
	b.WriteString("\n")

	return b.String()
}

func (claudeCodeTheme) RenderCode(info CodeInfo, width, height int) string {
	var b strings.Builder

	// Header bar
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(claudePurple)).
		Border(lipgloss.RoundedBorder(), true, false, true, false).
		BorderForeground(lipgloss.Color(claudePurple)).
		Width(width - 2)

	headerText := fmt.Sprintf(" ● Editing %s │ claude-fish", info.FileName)
	b.WriteString(headerStyle.Render(headerText))
	b.WriteString("\n")

	// AI preamble
	preambleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple))
	b.WriteString("✦ ")
	b.WriteString(preambleStyle.Render(fmt.Sprintf("Let me implement the changes in %s:", info.FileName)))
	b.WriteString("\n")

	// Code block
	if info.Displayed > 0 {
		codeContent := info.Content
		if info.Displayed < len(info.Content) {
			codeContent = info.Content[:info.Displayed]
		}

		codeStyle := lipgloss.NewStyle().
			Width(width - 4)

		fileLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeGray)).Render(
			fmt.Sprintf("┌─ %s", info.FileName))
		b.WriteString(fileLabel)
		b.WriteString("\n")
		b.WriteString(codeStyle.Render(codeContent))

		// Blink cursor at end
		if info.Displayed < info.Total {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(claudePurple)).Render("▌"))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (claudeCodeTheme) RenderStatusBar(info StatusInfo, width int) string {
	var parts []string
	switch info.Mode {
	case "welcome":
		parts = []string{"Space: start", "Q: quit"}
	case "reading":
		parts = []string{"Space: next", "B: back", "S: style", "Tab: boss", "Q: quit"}
	case "boss":
		parts = []string{"Tab: back to novel"}
	}

	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563")).Render(" │ ")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(claudeLightPurple))
	sepChar := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563")).Render(" | ")

	var styled []string
	for _, h := range info.Hints {
		styled = append(styled, hintStyle.Render(h.Key)+": "+h.Desc)
	}

	return sep + strings.Join(styled, sepChar) + sep
}
```

- [ ] **Step 4: Delete stub.go**

```bash
rm internal/theme/stub.go
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/theme/ -v -run TestClaudeCode
```

Expected: PASS

- [ ] **Step 6: Verify all packages still compile**

```bash
go build ./...
```

Expected: SUCCESS (Codex and OpenCode themes will fail — create temporary stubs)

- [ ] **Step 7: Create temporary stubs for remaining themes**

```go
// internal/theme/codex.go
package theme

import "github.com/charmbracelet/lipgloss"

type codexTheme struct{}

func NewCodex() Theme { return codexTheme{} }

func (codexTheme) Name() string                          { return "codex" }
func (codexTheme) AccentColor() lipgloss.Color           { return "#10a37f" }
func (codexTheme) UsableHeight(h int) int                { return h - 6 }
func (codexTheme) RenderWelcome(WelcomeInfo) string       { return "codex welcome\n" }
func (codexTheme) RenderPage(i PageInfo, _, _ int) string { return "codex page\n" }
func (codexTheme) RenderCode(i CodeInfo, _, _ int) string { return "codex code\n" }
func (codexTheme) RenderStatusBar(s StatusInfo, w int) string { return "codex status\n" }
```

```go
// internal/theme/opencode.go
package theme

import "github.com/charmbracelet/lipgloss"

type opencodeTheme struct{}

func NewOpenCode() Theme { return opencodeTheme{} }

func (opencodeTheme) Name() string                          { return "opencode" }
func (opencodeTheme) AccentColor() lipgloss.Color           { return "#89b4fa" }
func (opencodeTheme) UsableHeight(h int) int                { return h - 7 }
func (opencodeTheme) RenderWelcome(WelcomeInfo) string       { return "opencode welcome\n" }
func (opencodeTheme) RenderPage(i PageInfo, _, _ int) string { return "opencode page\n" }
func (opencodeTheme) RenderCode(i CodeInfo, _, _ int) string { return "opencode code\n" }
func (opencodeTheme) RenderStatusBar(s StatusInfo, w int) string { return "opencode status\n" }
```

- [ ] **Step 8: Verify compilation**

```bash
go build ./...
```

Expected: SUCCESS

- [ ] **Step 9: Commit**

```bash
git add internal/theme/
git commit -m "feat: implement Claude Code theme with full visual rendering"
```

---

### Task 8: Streamer

**Files:**
- Create: `internal/streamer.go`
- Create: `internal/streamer_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/streamer_test.go
package internal

import "testing"

func TestStreamerVisibleContent(t *testing.T) {
	s := NewStreamer("package main\n\nfunc main() {}\n", "main.go", 25)
	if s.FileName() != "main.go" {
		t.Errorf("FileName() = %q, want %q", s.FileName(), "main.go")
	}
	if s.Total() != len("package main\n\nfunc main() {}\n") {
		t.Errorf("Total() mismatch")
	}

	// Initially nothing displayed
	if s.Displayed() != 0 {
		t.Errorf("Displayed() = %d, want 0", s.Displayed())
	}

	// Advance
	s.Advance(5)
	if s.Displayed() != 5 {
		t.Errorf("after Advance(5), Displayed() = %d, want 5", s.Displayed())
	}

	// VisibleContent returns first 5 chars
	vis := s.VisibleContent()
	if vis != "packa" {
		t.Errorf("VisibleContent() = %q, want %q", vis, "packa")
	}
}

func TestStreamerLoop(t *testing.T) {
	code := "abc"
	s := NewStreamer(code, "test.go", 25)

	// Advance past end
	s.Advance(10)
	if !s.Done() {
		t.Error("expected Done() after advancing past total")
	}

	// Reset should loop
	s.Reset()
	if s.Displayed() != 0 {
		t.Errorf("after Reset(), Displayed() = %d, want 0", s.Displayed())
	}
}

func TestStreamerNeedsHint(t *testing.T) {
	s := NewStreamer("line1\nline2\nline3\nline4\nline5\n", "test.go", 25)

	// Should need hint every ~4 newlines
	s.Advance(5) // "line1"
	if !s.NeedsHint() {
		t.Error("expected NeedsHint after first newline content")
	}
	s.MarkHintShown()

	// Shouldn't need another hint immediately
	if s.NeedsHint() {
		t.Error("should not need hint immediately after showing one")
	}
}

func TestStreamerPreamble(t *testing.T) {
	s := NewStreamer("code", "handler.go", 25)
	preamble := s.Preamble()
	if preamble == "" {
		t.Error("Preamble() returned empty string")
	}
	if !contains(preamble, "handler.go") {
		t.Errorf("Preamble() = %q, should contain filename", preamble)
	}
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ -v -run TestStreamer
```

Expected: FAIL — `NewStreamer` not defined

- [ ] **Step 3: Write Streamer implementation**

```go
// internal/streamer.go
package internal

import (
	"fmt"
	"math/rand"
	"strings"
)

// Streamer manages the state of code streaming output.
type Streamer struct {
	code      string
	fileName  string
	speed     int // ms per character
	displayed int
	lastHintLine int // line count when last hint was shown
}

func NewStreamer(code, fileName string, speed int) *Streamer {
	return &Streamer{
		code:     code,
		fileName: fileName,
		speed:    speed,
	}
}

func (s *Streamer) FileName() string    { return s.fileName }
func (s *Streamer) Total() int          { return len(s.code) }
func (s *Streamer) Displayed() int      { return s.displayed }
func (s *Streamer) Speed() int          { return s.speed }

func (s *Streamer) Done() bool { return s.displayed >= len(s.code) }

func (s *Streamer) VisibleContent() string {
	if s.displayed >= len(s.code) {
		return s.code
	}
	return s.code[:s.displayed]
}

// Advance adds jitter and advances the displayed count by 1 character.
func (s *Streamer) Advance(count int) {
	s.displayed += count
	if s.displayed > len(s.code) {
		s.displayed = len(s.code)
	}
}

func (s *Streamer) Reset() {
	s.displayed = 0
	s.lastHintLine = 0
}

// NeedsHint returns true when a "tool call" hint should be inserted.
// Triggers every ~4 newlines in the visible content.
func (s *Streamer) NeedsHint() bool {
	visible := s.VisibleContent()
	currentLines := strings.Count(visible, "\n")
	if currentLines > 0 && currentLines-s.lastHintLine >= 4 {
		return true
	}
	return false
}

func (s *Streamer) MarkHintShown() {
	visible := s.VisibleContent()
	s.lastHintLine = strings.Count(visible, "\n")
}

// Preamble returns the "AI thinking" text shown before code streams.
func (s *Streamer) Preamble() string {
	return fmt.Sprintf("✦ Let me implement the changes in %s:", s.fileName)
}

// JitterSpeed returns speed with random jitter (±40%).
func (s *Streamer) JitterSpeed() int {
	jitter := float64(s.speed) * 0.4
	return s.speed + int(rand.Float64()*jitter*2-jitter)
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/ -v -run TestStreamer
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/streamer.go internal/streamer_test.go
git commit -m "feat: add code streaming engine with hint insertion and jitter"
```

---

### Task 9: Boss Mode State

**Files:**
- Create: `internal/boss.go`
- Create: `internal/boss_test.go`

- [ ] **Step 1: Write the failing test**

```go
// internal/boss_test.go
package internal

import "testing"

func TestBossModeInit(t *testing.T) {
	b := NewBossMode("package main\n", "main.go", 25)
	if b.Active() {
		t.Error("new BossMode should not be active")
	}
	if b.HasCode() != true {
		t.Error("HasCode() should be true when code is loaded")
	}
}

func TestBossModeNoCode(t *testing.T) {
	b := NewBossMode("", "", 25)
	if b.HasCode() {
		t.Error("HasCode() should be false when no code")
	}
}

func TestBossModeActivateDeactivate(t *testing.T) {
	b := NewBossMode("code", "test.go", 25)

	b.Activate()
	if !b.Active() {
		t.Error("should be active after Activate()")
	}
	if b.Streamer().Displayed() != 0 {
		t.Error("streamer should start at 0")
	}

	// Simulate some streaming
	b.Streamer().Advance(2)
	if b.Streamer().Displayed() != 2 {
		t.Error("streamer should have advanced")
	}

	b.Deactivate()
	if b.Active() {
		t.Error("should not be active after Deactivate()")
	}

	// Reactivate — should resume from where it left off
	b.Activate()
	if b.Streamer().Displayed() != 2 {
		t.Error("streamer should resume from 2, got %d", b.Streamer().Displayed())
	}

	// Stream to end
	b.Streamer().Advance(100)
	if !b.Streamer().Done() {
		t.Error("streamer should be done")
	}

	// Deactivate and reactivate — should loop
	b.Deactivate()
	b.Activate()
	if b.Streamer().Displayed() != 0 {
		t.Error("after loop, streamer should start from 0")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ -v -run TestBossMode
```

Expected: FAIL — `NewBossMode` not defined

- [ ] **Step 3: Write Boss Mode implementation**

```go
// internal/boss.go
package internal

// BossMode manages the boss-key state: instant switching between
// novel reading and fake code editing.
type BossMode struct {
	streamer *Streamer
	active   bool
	hasCode  bool
}

func NewBossMode(code, fileName string, speed int) *BossMode {
	return &BossMode{
		streamer: NewStreamer(code, fileName, speed),
		hasCode:  code != "",
	}
}

func (b *BossMode) Active() bool    { return b.active }
func (b *BossMode) HasCode() bool   { return b.hasCode }
func (b *BossMode) Streamer() *Streamer { return b.streamer }

func (b *BossMode) Activate() {
	b.active = true
	// If code is done, loop from beginning
	if b.streamer.Done() {
		b.streamer.Reset()
	}
}

func (b *BossMode) Deactivate() {
	b.active = false
	// Don't reset — resume from where we left off next time
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/ -v -run TestBossMode
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/boss.go internal/boss_test.go
git commit -m "feat: add BossMode state with activate/deactivate and loop"
```

---

### Task 10: Bubble Tea App Core

**Files:**
- Create: `internal/app.go`
- Create: `internal/app_test.go`

This is the heart of the application — the Bubble Tea model with state machine.

- [ ] **Step 1: Write the failing test for state transitions**

```go
// internal/app_test.go
package internal

import (
	"testing"

	"claude-fish/internal/reader"
	"claude-fish/internal/theme"
)

func newTestModel() model {
	r := &reader.TXTReader{}
	r.Load("../../testdata/sample.txt")
	th := theme.NewClaudeCode()
	return newModel(r, th, nil, 80, 24, 25)
}

func TestInitialState(t *testing.T) {
	m := newTestModel()
	if m.state != stateWelcome {
		t.Errorf("initial state = %v, want %v", m.state, stateWelcome)
	}
}

func TestWelcomeToReading(t *testing.T) {
	m := newTestModel()
	// Simulate Space key press
	msg := keyMsg(' ')
	newModel, _ := m.Update(msg)
	m2 := newModel.(model)
	if m2.state != stateReading {
		t.Errorf("after Space, state = %v, want %v", m2.state, stateReading)
	}
}

func TestReadingToBossAndBack(t *testing.T) {
	m := newTestModel()
	// Enter reading mode
	m, _ = m.Update(keyMsg(' '))
	m = m.(model)

	// Need code for boss mode
	m.boss = NewBossMode("code", "test.go", 25)

	// Tab to boss mode
	m, _ = m.Update(tabKeyMsg())
	m = m.(model)
	if m.state != stateBoss {
		t.Errorf("after Tab, state = %v, want %v", m.state, stateBoss)
	}

	// Tab back to reading
	m, _ = m.Update(tabKeyMsg())
	m = m.(model)
	if m.state != stateReading {
		t.Errorf("after Tab back, state = %v, want %v", m.state, stateReading)
	}
}

func TestReadingPageNavigation(t *testing.T) {
	m := newTestModel()
	m, _ = m.Update(keyMsg(' '))
	m = m.(model)

	if m.pager.Page() != 0 {
		t.Errorf("initial page = %d, want 0", m.pager.Page())
	}

	// Space for next page (might have only 1 page, so check)
	totalPg := m.pager.TotalPages()
	if totalPg > 1 {
		m, _ = m.Update(keyMsg(' '))
		m = m.(model)
		if m.pager.Page() != 1 {
			t.Errorf("after Space, page = %d, want 1", m.pager.Page())
		}
	}
}

// Helper: create a simple key message.
// In Bubble Tea, tea.KeyMsg contains a tea.Key with Type and Runes.
// For testing we use a simpler approach.
type testKey struct {
	r    rune
	tab  bool
}

func keyMsg(r rune) testKey { return testKey{r: r} }
func tabKeyMsg() testKey    { return testKey{tab: true} }
```

- [ ] **Step 2: Write the App model**

Note: We use a custom key handling approach that wraps Bubble Tea's key types to make testing easier.

```go
// internal/app.go
package internal

import (
	"fmt"
	"os"
	"strings"

	"claude-fish/internal/reader"
	"claude-fish/internal/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type appState int

const (
	stateWelcome appState = iota
	stateReading
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
	err        error
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

	return model{
		state:      stateWelcome,
		theme:      th,
		themes:     themes,
		themeIndex: themeIndex,
		pager:      pg,
		boss:       boss,
		width:      width,
		height:     height,
		speed:      speed,
	}
}

// NewApp creates the Bubble Tea program.
func NewApp(r reader.Reader, th theme.Theme, code, codeFile string, speed int) *tea.Program {
	m := newModel(r, th, NewBossMode(code, codeFile, speed), 80, 24, speed)
	if r != nil {
		chapters := r.Chapters()
		if len(chapters) > 0 {
			m.fileName = "" // set from caller
		}
	}
	return tea.NewProgram(m, tea.WithAltScreen())
}

// SetFileName sets the display filename.
func (m *model) SetFileName(name string) {
	m.fileName = name
}

func (model) Init() tea.Cmd { return nil }

type tickMsg time.Time

// Update handles Bubble Tea messages. For testability, we also accept
// our testKey type. In production, Bubble Tea sends tea.KeyMsg.
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

	case testKey:
		return m.handleTestKey(msg)
	}

	return m, nil
}

func (m model) handleKey(key string) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateWelcome:
		switch key {
		case " ", "enter":
			m.state = stateReading
		case "q", "ctrl+c":
			return m, tea.Quit
		}

	case stateReading:
		switch key {
		case " ", "right":
			m.pager.NextPage()
		case "b", "left":
			m.pager.PrevPage()
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

func (m model) handleTestKey(k testKey) (tea.Model, tea.Cmd) {
	if k.tab {
		return m.handleKey("tab")
	}
	return m.handleKey(string(k.r))
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n", m.err)
	}

	switch m.state {
	case stateWelcome:
		chapters := 0
		if m.pager != nil {
			chapters = len(m.pager.Chapters())
		}
		return m.theme.RenderWelcome(theme.WelcomeInfo{
			Version:   "v1.0.0",
			FileName:  m.fileName,
			Chapters:  chapters,
			ThemeName: m.theme.Name(),
		}) + "\n" + renderSeparator(m.width) + "\n❯ \n" + renderSeparator(m.width)

	case stateReading:
		content := ""
		if m.pager != nil {
			content = m.pager.CurrentContent()
		}
		return m.theme.RenderPage(theme.PageInfo{
			ChapterTitle: m.pager.CurrentTitle(),
			Content:      content,
			PageNum:      m.pager.Page(),
			TotalPages:   m.pager.TotalPages(),
			FileName:     m.fileName,
		}, m.width, m.height) + "\n" + renderSeparator(m.width) + "\n❯ \n" + renderSeparator(m.width)

	case stateBoss:
		s := m.boss.Streamer()
		return m.theme.RenderCode(theme.CodeInfo{
			FileName:  s.FileName(),
			Content:   s.VisibleContent(),
			Displayed: s.Displayed(),
			Total:     s.Total(),
		}, m.width, m.height) + "\n" + renderSeparator(m.width) + "\n❯ \n" + renderSeparator(m.width)
	}

	return ""
}

func renderSeparator(width int) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4b5563")).
		Render(strings.Repeat("─", width))
}
```

We need to add the `time` import and fix the `tickMsg` type definition:

```go
// Add at top of app.go with imports:
import "time"

// tickMsg needs to be defined as:
type tickMsg time.Time
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./...
```

Expected: May need minor import fixes. Ensure `time` and `os` are in imports, and the `fmt` for `os` reference is correct.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/ -v -run Test
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/app.go internal/app_test.go
git commit -m "feat: add Bubble Tea app with state machine and key handling"
```

---

### Task 11: Codex Theme

**Files:**
- Modify: `internal/theme/codex.go` (replace stub)

- [ ] **Step 1: Write the failing test**

```go
// internal/theme/codex_test.go
package theme

import (
	"strings"
	"testing"
)

func TestCodexName(t *testing.T) {
	th := NewCodex()
	if th.Name() != "codex" {
		t.Errorf("Name() = %q, want %q", th.Name(), "codex")
	}
}

func TestCodexRenderPage(t *testing.T) {
	th := NewCodex()
	info := PageInfo{
		ChapterTitle: "Chapter 1",
		Content:      "Test content.",
		PageNum:      0,
		TotalPages:   5,
		FileName:     "test.txt",
	}
	output := th.RenderPage(info, 80, 24)

	if !strings.Contains(output, "Chapter 1") {
		t.Error("missing chapter title")
	}
	if !strings.Contains(output, "Test content") {
		t.Error("missing content")
	}
}

func TestCodexRenderWelcome(t *testing.T) {
	th := NewCodex()
	output := th.RenderWelcome(WelcomeInfo{Version: "v1.0", FileName: "test.txt", Chapters: 5})
	if !strings.Contains(output, "test.txt") {
		t.Error("missing filename")
	}
}

func TestCodexRenderCode(t *testing.T) {
	th := NewCodex()
	output := th.RenderCode(CodeInfo{FileName: "main.go", Content: "package main\n", Displayed: 8}, 80, 24)
	if !strings.Contains(output, "main.go") {
		t.Error("missing filename")
	}
}
```

- [ ] **Step 2: Replace codex.go stub with full implementation**

```go
// internal/theme/codex.go
package theme

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	codexGreen   = "#10a37f"
	codexDark    = "#0d1117"
	codexGray    = "#6b7280"
	codexYellow  = "#f59e0b"
	codexWhite   = "#e0e0e0"
)

type codexTheme struct{}

func NewCodex() Theme { return codexTheme{} }

func (codexTheme) Name() string              { return "codex" }
func (codexTheme) AccentColor() lipgloss.Color { return codexGreen }
func (codexTheme) UsableHeight(h int) int    { return h - 6 }

func (codexTheme) RenderWelcome(info WelcomeInfo) string {
	var b strings.Builder
	green := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGreen)).Bold(true)
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGray))

	b.WriteString(green.Render("codex"))
	b.WriteString(" ")
	b.WriteString(gray.Render(info.Version))
	b.WriteString("\n")
	b.WriteString(gray.Render(strings.Repeat("─", 60)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Loaded: %s (%d chapters)\n", info.FileName, info.Chapters))
	b.WriteString("\n")
	b.WriteString(gray.Render("Press Space to start reading"))
	b.WriteString("\n")
	return b.String()
}

func (codexTheme) RenderPage(info PageInfo, width, _ int) string {
	var b strings.Builder
	green := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGreen))
	white := lipgloss.NewStyle().Foreground(lipgloss.Color(codexWhite))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGray))

	// Title with > prompt
	b.WriteString(green.Render(">"))
	b.WriteString(" ")
	b.WriteString(white.Render(info.ChapterTitle))
	b.WriteString("\n\n")

	// Content
	b.WriteString(white.Render(info.Content))
	b.WriteString("\n\n")

	// Progress bar
	progress := float64(info.PageNum+1) / float64(info.TotalPages)
	barWidth := 30
	filled := int(progress * float64(barWidth))
	bar := green.Render(strings.Repeat("█", filled)) + gray.Render(strings.Repeat("░", barWidth-filled))
	b.WriteString(fmt.Sprintf("[%s] %d/%d", bar, info.PageNum+1, info.TotalPages))
	b.WriteString("\n")

	return b.String()
}

func (codexTheme) RenderCode(info CodeInfo, width, _ int) string {
	var b strings.Builder
	green := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGreen))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGray))
	white := lipgloss.NewStyle().Foreground(lipgloss.Color(codexWhite))

	b.WriteString(green.Render(">"))
	b.WriteString(" ")
	b.WriteString(white.Render(fmt.Sprintf("Editing %s", info.FileName)))
	b.WriteString("\n\n")

	if info.Displayed > 0 {
		visible := info.Content
		if info.Displayed < len(info.Content) {
			visible = info.Content[:info.Displayed]
		}
		b.WriteString(gray.Render(fmt.Sprintf("┌─ %s", info.FileName)))
		b.WriteString("\n")
		b.WriteString(white.Render(visible))
		if info.Displayed < info.Total {
			b.WriteString(green.Render("▌"))
		}
		b.WriteString("\n")
	}

	b.WriteString(gray.Render("Writing..."))
	b.WriteString("\n")

	return b.String()
}

func (codexTheme) RenderStatusBar(info StatusInfo, width int) string {
	var parts []string
	switch info.Mode {
	case "welcome":
		parts = []string{"Space: start", "q: quit"}
	case "reading":
		parts = []string{"Space: next", "q: quit", "!: boss"}
	case "boss":
		parts = []string{"Tab: back"}
	}

	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#4b5563")).Render(" | ")
	green := lipgloss.NewStyle().Foreground(lipgloss.Color(codexGreen))

	return green.Render("→") + sep + strings.Join(parts, sep)
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/theme/ -v -run TestCodex
```

Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/theme/codex.go internal/theme/codex_test.go
git commit -m "feat: implement Codex CLI theme"
```

---

### Task 12: opencode Theme

**Files:**
- Modify: `internal/theme/opencode.go` (replace stub)

- [ ] **Step 1: Write the failing test**

```go
// internal/theme/opencode_test.go
package theme

import (
	"strings"
	"testing"
)

func TestOpenCodeName(t *testing.T) {
	th := NewOpenCode()
	if th.Name() != "opencode" {
		t.Errorf("Name() = %q, want %q", th.Name(), "opencode")
	}
}

func TestOpenCodeRenderPage(t *testing.T) {
	th := NewOpenCode()
	info := PageInfo{
		ChapterTitle: "第一章",
		Content:      "测试内容。",
		PageNum:      0,
		TotalPages:   3,
		FileName:     "test.txt",
	}
	output := th.RenderPage(info, 80, 24)
	if !strings.Contains(output, "第一章") {
		t.Error("missing chapter title")
	}
	if !strings.Contains(output, "测试内容") {
		t.Error("missing content")
	}
}

func TestOpenCodeRenderWelcome(t *testing.T) {
	th := NewOpenCode()
	output := th.RenderWelcome(WelcomeInfo{FileName: "test.txt", Chapters: 3})
	if !strings.Contains(output, "test.txt") {
		t.Error("missing filename")
	}
}

func TestOpenCodeRenderCode(t *testing.T) {
	th := NewOpenCode()
	output := th.RenderCode(CodeInfo{FileName: "main.go", Content: "code", Displayed: 4}, 80, 24)
	if !strings.Contains(output, "main.go") {
		t.Error("missing filename")
	}
}
```

- [ ] **Step 2: Replace opencode.go stub with full implementation**

```go
// internal/theme/opencode.go
package theme

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	ocBlue      = "#89b4fa"
	ocText      = "#cdd6f4"
	ocSubtext   = "#6c7086"
	ocSurface   = "#1e1e2e"
	ocOverlay   = "#45475a"
	ocRed       = "#f38ba8"
)

type opencodeTheme struct{}

func NewOpenCode() Theme { return opencodeTheme{} }

func (opencodeTheme) Name() string              { return "opencode" }
func (opencodeTheme) AccentColor() lipgloss.Color { return ocBlue }
func (opencodeTheme) UsableHeight(h int) int    { return h - 7 }

func (opencodeTheme) RenderWelcome(info WelcomeInfo) string {
	var b strings.Builder
	blue := lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue)).Bold(true)
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))

	// Tab bar
	b.WriteString(renderTabBar("Welcome", []string{"Welcome", "Files", "Config"}))
	b.WriteString("\n")

	// Content box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ocOverlay)).
		Width(60).
		Padding(1)

	content := blue.Render("opencode") + "\n\n" +
		fmt.Sprintf("Loaded: %s (%d chapters)\n", info.FileName, info.Chapters) +
		gray.Render("Press Space to start")
	b.WriteString(box.Render(content))
	b.WriteString("\n")
	return b.String()
}

func (opencodeTheme) RenderPage(info PageInfo, width, _ int) string {
	var b strings.Builder

	// Tab bar
	b.WriteString(renderTabBar("Chat", []string{"Chat", "Files", "Diff"}))
	b.WriteString("\n")

	// Chat bubble
	blue := lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue))
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(ocText))

	b.WriteString(blue.Render("assistant"))
	b.WriteString("\n")
	b.WriteString(text.Render("  "+info.ChapterTitle))
	b.WriteString("\n\n")
	b.WriteString(text.Render(info.Content))
	b.WriteString("\n\n")

	// Footer
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))
	b.WriteString(gray.Render(fmt.Sprintf("%d/%d pages", info.PageNum+1, info.TotalPages)))
	b.WriteString("\n")

	return b.String()
}

func (opencodeTheme) RenderCode(info CodeInfo, width, _ int) string {
	var b strings.Builder

	b.WriteString(renderTabBar("Diff", []string{"Chat", "Files", "Diff"}))
	b.WriteString("\n")

	blue := lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue))
	text := lipgloss.NewStyle().Foreground(lipgloss.Color(ocText))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))

	b.WriteString(blue.Render("assistant"))
	b.WriteString("\n")

	if info.Displayed > 0 {
		visible := info.Content
		if info.Displayed < len(info.Content) {
			visible = info.Content[:info.Displayed]
		}
		b.WriteString(gray.Render(fmt.Sprintf("┌─ %s", info.FileName)))
		b.WriteString("\n")
		b.WriteString(text.Render(visible))
		if info.Displayed < info.Total {
			b.WriteString(blue.Render("▌"))
		}
		b.WriteString("\n")
	}

	b.WriteString(gray.Render("Writing..."))
	b.WriteString("\n")
	return b.String()
}

func (opencodeTheme) RenderStatusBar(info StatusInfo, width int) string {
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))
	red := lipgloss.NewStyle().Foreground(lipgloss.Color(ocRed))

	switch info.Mode {
	case "welcome":
		return gray.Render("Space: start | q: quit")
	case "reading":
		return gray.Render("Space: next | B: back | S: style | ") + red.Render("Tab") + gray.Render(": boss")
	case "boss":
		return red.Render("[BOSS MODE] ") + gray.Render("Tab: back")
	}
	return ""
}

func renderTabBar(active string, tabs []string) string {
	var b strings.Builder
	blue := lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue))
	gray := lipgloss.NewStyle().Foreground(lipgloss.Color(ocSubtext))

	for _, tab := range tabs {
		if tab == active {
			b.WriteString(blue.Bold(true).Render(tab))
			b.WriteString("  ")
		} else {
			b.WriteString(gray.Render(tab))
			b.WriteString("  ")
		}
	}
	// Underline for active tab
	b.WriteString("\n")
	for _, tab := range tabs {
		if tab == active {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ocBlue)).Render(strings.Repeat("─", len(tab))))
			b.WriteString("  ")
		} else {
			b.WriteString(strings.Repeat(" ", len(tab)))
			b.WriteString("  ")
		}
	}
	return b.String()
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./internal/theme/ -v -run TestOpenCode
```

Expected: PASS

- [ ] **Step 4: Run all theme tests**

```bash
go test ./internal/theme/ -v
```

Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add internal/theme/opencode.go internal/theme/opencode_test.go
git commit -m "feat: implement opencode theme with Catppuccin palette"
```

---

### Task 13: CLI with Cobra

**Files:**
- Create: `cmd/root.go`
- Modify: `main.go`

- [ ] **Step 1: Write cmd/root.go**

```go
// cmd/root.go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"claude-fish/internal"
	"claude-fish/internal/reader"
	"claude-fish/internal/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var (
	codeFile string
	themeName string
	speed    int
)

var rootCmd = &cobra.Command{
	Use:   "claude-fish <novel-file>",
	Short: "A terminal novel reader disguised as a CLI coding tool",
	Args:  cobra.ExactArgs(1),
	Run:   run,
}

func init() {
	rootCmd.Flags().StringVarP(&codeFile, "code", "c", "", "code file for boss mode")
	rootCmd.Flags().StringVarP(&themeName, "theme", "t", "claude", "theme: claude, codex, opencode")
	rootCmd.Flags().IntVar(&speed, "speed", 25, "streaming speed in ms/char")
}

func run(cmd *cobra.Command, args []string) {
	novelPath := args[0]

	// Detect format and create reader
	r := newReader(novelPath)
	if r == nil {
		fmt.Fprintf(os.Stderr, "Unsupported file format: %s\n", filepath.Ext(novelPath))
		os.Exit(1)
	}

	if err := r.Load(novelPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", novelPath, err)
		os.Exit(1)
	}

	// Load code file for boss mode
	var code, cfName string
	if codeFile != "" {
		data, err := os.ReadFile(codeFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading code file %s: %v\n", codeFile, err)
			os.Exit(1)
		}
		code = string(data)
		cfName = filepath.Base(codeFile)
	}

	// Select theme
	th := theme.FindByName(themeName)

	// Create and run app
	p := internal.NewApp(r, th, code, cfName, speed)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newReader(path string) reader.Reader {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt":
		return &reader.TXTReader{}
	case ".md", ".markdown":
		return &reader.MarkdownReader{}
	case ".epub":
		return &reader.EPUBReader{}
	default:
		return nil
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 2: Write main.go**

```go
// main.go
package main

import "claude-fish/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 3: Verify compilation**

```bash
go build -o claude-fish .
```

Expected: SUCCESS — binary `claude-fish` created

- [ ] **Step 4: Test CLI help**

```bash
./claude-fish --help
```

Expected: Shows usage with flags

- [ ] **Step 5: Commit**

```bash
git add cmd/root.go main.go
git commit -m "feat: add CLI with cobra, format auto-detection, and flag parsing"
```

---

### Task 14: Integration Test & Syntax Highlighting

**Files:**
- Create: `internal/highlight.go`
- Modify: `internal/theme/claudecode.go` — integrate syntax highlighting into code rendering
- Modify: `internal/theme/codex.go` — integrate syntax highlighting
- Modify: `internal/theme/opencode.go` — integrate syntax highlighting

- [ ] **Step 1: Write syntax highlighting helper**

```go
// internal/highlight.go
package internal

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromaStyles "github.com/alecthomas/chroma/v2/styles"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/formatters"
)

// HighlightCode applies syntax highlighting to code, returning ANSI-colored text.
func HighlightCode(code, fileName string) string {
	lexer := lexers.Match(fileName)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := chromaStyles.Get("dracula")
	if style == nil {
		style = chromaStyles.Fallback
	}

	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return code
	}

	var b strings.Builder
	formatter.Format(&b, style, iterator)
	return b.String()
}
```

- [ ] **Step 2: Update each theme's RenderCode to use HighlightCode**

In each theme file, update the `RenderCode` method to highlight the visible content before rendering. Add to the code block section:

```go
// In the RenderCode method, before rendering code content:
highlighted := internal.HighlightCode(visible, info.FileName)
// Then render `highlighted` instead of `visible`
```

Note: This creates an import dependency from theme → internal. To avoid circular imports, move `highlight.go` to a separate package or pass pre-highlighted content. The cleanest approach: make `CodeInfo.Content` already highlighted before it reaches the theme. Update `app.go` View() for boss mode:

```go
// In app.go View(), boss mode section:
highlighted := HighlightCode(s.VisibleContent(), s.FileName())
return m.theme.RenderCode(theme.CodeInfo{
    FileName:  s.FileName(),
    Content:   highlighted,  // pre-highlighted
    Displayed: s.Displayed(),
    Total:     s.Total(),
}, m.width, m.height)
```

This way themes just render the content as-is, and syntax highlighting is done in the app layer.

- [ ] **Step 3: Verify compilation and test**

```bash
go build -o claude-fish .
```

Expected: SUCCESS

- [ ] **Step 4: Manual integration test**

```bash
# Test with TXT file
./claude-fish testdata/sample.txt -c main.go

# Test with Markdown file
./claude-fish testdata/sample.md -t codex

# Test theme switching: press S in reading mode
# Test boss mode: press Tab
# Test boss mode back: press Tab again
```

Expected: Full TUI with welcome screen, reading mode, boss mode switching

- [ ] **Step 5: Commit**

```bash
git add internal/highlight.go
git add -u
git commit -m "feat: add syntax highlighting and integration wiring"
```

---

## Self-Review Checklist

- **Spec coverage:** All requirements mapped to tasks: Theme interface (T2, T7, T11, T12), Reader (T3-T5), Pager (T6), Streamer (T8), BossMode (T9), App (T10), CLI (T13), Highlighting (T14)
- **Placeholder scan:** No TBD/TODO found. All code steps include actual implementations.
- **Type consistency:** `Reader` interface methods match across TXT/Markdown/EPUB implementations. `Theme` interface methods match across all three theme implementations. `model` struct fields used consistently in app.go.
