package internal

import (
	"fmt"
	"math/rand"
	"strings"
)

type Streamer struct {
	code         string
	fileName     string
	speed        int
	displayed    int
	lastHintLine int
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
func (s *Streamer) Done() bool          { return s.displayed >= len(s.code) }

func (s *Streamer) VisibleContent() string {
	if s.displayed >= len(s.code) {
		return s.code
	}
	return s.code[:s.displayed]
}

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

func (s *Streamer) Preamble() string {
	return fmt.Sprintf("✦ Let me implement the changes in %s:", s.fileName)
}

func (s *Streamer) JitterSpeed() int {
	jitter := float64(s.speed) * 0.4
	return s.speed + int(rand.Float64()*jitter*2-jitter)
}
