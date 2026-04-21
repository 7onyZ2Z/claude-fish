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
		ChapterTitle:  "Chapter 1",
		Content:       "Test content.",
		PageNum:       0,
		TotalPages:    5,
		FileName:      "test.txt",
		TotalChapters: 10,
		ThemeName:     "codex",
		Version:       "v1.0",
	}
	output := th.RenderPage(info, 80, 24)
	if !strings.Contains(output, "Chapter 1") {
		t.Error("missing chapter title")
	}
	if !strings.Contains(output, "Test content") {
		t.Error("missing content")
	}
	if !strings.Contains(output, "test.txt") {
		t.Error("missing filename in header")
	}
}

func TestCodexRenderCode(t *testing.T) {
	th := NewCodex()
	output := th.RenderCode(CodeInfo{
		FileName:  "main.go",
		Content:   "package main\n",
		Displayed: 8,
		ThemeName: "codex",
		Version:   "v1.0",
	}, 80, 24)
	if !strings.Contains(output, "main.go") {
		t.Error("missing filename")
	}
}
