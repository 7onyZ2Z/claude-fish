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
