package internal

import (
	"strings"
	"testing"
)

func TestStreamerVisibleContent(t *testing.T) {
	s := NewStreamer("package main\n\nfunc main() {}\n", "main.go", 25)
	if s.FileName() != "main.go" {
		t.Errorf("FileName() = %q, want %q", s.FileName(), "main.go")
	}

	s.Advance(5)
	if s.Displayed() != 5 {
		t.Errorf("after Advance(5), Displayed() = %d, want 5", s.Displayed())
	}

	vis := s.VisibleContent()
	if vis != "packa" {
		t.Errorf("VisibleContent() = %q, want %q", vis, "packa")
	}
}

func TestStreamerLoop(t *testing.T) {
	code := "abc"
	s := NewStreamer(code, "test.go", 25)

	s.Advance(10)
	if !s.Done() {
		t.Error("expected Done() after advancing past total")
	}

	s.Reset()
	if s.Displayed() != 0 {
		t.Errorf("after Reset(), Displayed() = %d, want 0", s.Displayed())
	}
}

func TestStreamerPreamble(t *testing.T) {
	s := NewStreamer("code", "handler.go", 25)
	preamble := s.Preamble()
	if preamble == "" {
		t.Error("Preamble() returned empty string")
	}
	if !strings.Contains(preamble, "handler.go") {
		t.Errorf("Preamble() = %q, should contain filename", preamble)
	}
}
