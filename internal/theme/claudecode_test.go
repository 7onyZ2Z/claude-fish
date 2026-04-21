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
		ChapterTitle:  "第一章 开始",
		Content:       "这是测试内容。",
		PageNum:       1,
		TotalPages:    10,
		FileName:      "三体.txt",
		TotalChapters: 42,
		ThemeName:     "claude",
		Version:       "v2.1.116",
	}
	output := th.RenderPage(info, 80, 24)

	if !strings.Contains(output, "Claude Code") {
		t.Error("RenderPage output missing Claude Code branding")
	}
	if !strings.Contains(output, "Welcome back") {
		t.Error("RenderPage output missing Welcome back")
	}
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
		ThemeName: "claude",
		Version:   "v2.1.116",
	}
	output := th.RenderCode(info, 80, 24)

	if !strings.Contains(output, "Claude Code") {
		t.Error("RenderCode output missing Claude Code branding")
	}
	if !strings.Contains(output, "package main") {
		t.Error("RenderCode output missing content")
	}
	if !strings.Contains(output, "Claude Code") {
		t.Error("RenderCode output missing Claude Code branding")
	}
}

func TestClaudeCodeUsableHeight(t *testing.T) {
	th := NewClaudeCode()
	usable := th.UsableHeight(24)
	if usable <= 0 || usable >= 24 {
		t.Errorf("UsableHeight(24) = %d, want between 1 and 23", usable)
	}
}
