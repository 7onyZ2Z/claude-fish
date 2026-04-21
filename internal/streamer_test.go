package internal

import (
	"strings"
	"testing"
)

func TestStreamerVisibleContent(t *testing.T) {
	segs := []Segment{TextSegment("package main\n\nfunc main() {}\n")}
	s := NewStreamer(segs, 25)

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
	segs := []Segment{TextSegment("abc")}
	s := NewStreamer(segs, 25)

	s.Advance(10)
	if !s.Done() {
		t.Error("expected Done() after advancing past total")
	}

	s.Reset()
	if s.Displayed() != 0 {
		t.Errorf("after Reset(), Displayed() = %d, want 0", s.Displayed())
	}
}

func TestStreamerMultiSegment(t *testing.T) {
	segs := []Segment{
		ThinkSegment("thinking..."),
		TextSegment("done"),
	}
	s := NewStreamer(segs, 25)

	s.Advance(5)
	vis := s.VisibleContent()
	if !strings.Contains(vis, "think") {
		t.Errorf("VisibleContent() = %q, should contain 'think'", vis)
	}

	s.Advance(100)
	vis = s.VisibleContent()
	if !strings.Contains(vis, "done") {
		t.Errorf("VisibleContent() = %q, should contain 'done'", vis)
	}
}
