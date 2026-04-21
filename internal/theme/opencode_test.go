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
		ChapterTitle:  "第一章",
		Content:       "测试内容。",
		PageNum:       0,
		TotalPages:    3,
		FileName:      "test.txt",
		TotalChapters: 5,
		ThemeName:     "opencode",
		Version:       "v1.0",
	}
	output := th.RenderPage(info, 80, 24)
	if !strings.Contains(output, "第一章") {
		t.Error("missing chapter title")
	}
	if !strings.Contains(output, "测试内容") {
		t.Error("missing content")
	}
	if !strings.Contains(output, "test.txt") {
		t.Error("missing filename in header")
	}
}

func TestOpenCodeRenderCode(t *testing.T) {
	th := NewOpenCode()
	output := th.RenderCode(CodeInfo{
		FileName:  "main.go",
		Content:   "code",
		Displayed: 4,
		ThemeName: "opencode",
		Version:   "v1.0",
	}, 80, 24)
	if !strings.Contains(output, "main.go") {
		t.Error("missing filename")
	}
}
