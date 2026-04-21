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
	if !strings.Contains(output, "2/10") {
		t.Error("RenderPage output missing page indicator 2/10")
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
