package internal

import (
	"fmt"
	"math/rand"
	"strings"
)

type Streamer struct {
	segments []Segment
	segIdx   int
	pos      int
	speed    int
}

func NewStreamer(segments []Segment, speed int) *Streamer {
	if len(segments) == 0 {
		segments = []Segment{
			ThinkSegment("Initializing..."),
			TextSegment("Ready."),
		}
	}
	return &Streamer{
		segments: segments,
		speed:    speed,
	}
}

func (s *Streamer) FileName() string { return "output" }
func (s *Streamer) Total() int {
	total := 0
	for _, seg := range s.segments {
		total += len(seg.Content)
	}
	return total
}
func (s *Streamer) Done() bool { return s.segIdx >= len(s.segments) }

func (s *Streamer) Advance(count int) {
	if s.Done() {
		return
	}
	s.pos += count
	for s.pos > len(s.segments[s.segIdx].Content) && !s.Done() {
		s.pos -= len(s.segments[s.segIdx].Content)
		s.segIdx++
		if s.segIdx >= len(s.segments) {
			s.segIdx = len(s.segments)
			s.pos = 0
			return
		}
	}
	if s.segIdx < len(s.segments) && s.pos > len(s.segments[s.segIdx].Content) {
		s.pos = len(s.segments[s.segIdx].Content)
	}
}

func (s *Streamer) Reset() {
	s.segIdx = 0
	s.pos = 0
}

func (s *Streamer) Speed() int { return s.speed }

// ChunkSize returns a variable number of characters to advance per tick.
func (s *Streamer) ChunkSize() int {
	seg := s.CurrentSegmentType()
	switch seg {
	case SegmentThink:
		n := 3 + rand.Intn(8)
		if rand.Float64() < 0.15 {
			n += rand.Intn(12)
		}
		return n
	case SegmentText:
		n := 2 + rand.Intn(5)
		if rand.Float64() < 0.1 {
			n = 1
		}
		return n
	case SegmentCode:
		n := 2 + rand.Intn(4)
		if rand.Float64() < 0.25 {
			n += rand.Intn(6)
		}
		return n
	}
	return 3
}

// JitterSpeed returns tick interval with segment-aware variation.
func (s *Streamer) JitterSpeed() int {
	seg := s.CurrentSegmentType()
	switch seg {
	case SegmentThink:
		return 8 + rand.Intn(10)
	case SegmentText:
		if rand.Float64() < 0.08 {
			return 60 + rand.Intn(60)
		}
		return 12 + rand.Intn(18)
	case SegmentCode:
		if rand.Float64() < 0.05 {
			return 50 + rand.Intn(40)
		}
		return 10 + rand.Intn(12)
	}
	return 15
}

// VisibleContent returns the full rendered output up to current position.
func (s *Streamer) VisibleContent() string {
	var b strings.Builder
	for i := 0; i < s.segIdx && i < len(s.segments); i++ {
		s.renderSegment(&b, s.segments[i], -1)
	}
	if s.segIdx < len(s.segments) {
		s.renderSegment(&b, s.segments[s.segIdx], s.pos)
	}
	return b.String()
}

func (s *Streamer) renderSegment(b *strings.Builder, seg Segment, charLimit int) {
	content := seg.Content
	if charLimit >= 0 && charLimit < len(content) {
		content = content[:charLimit]
	}
	if content == "" {
		return
	}

	switch seg.Type {
	case SegmentThink:
		b.WriteString("\033[38;5;243m")
		b.WriteString(content)
		b.WriteString("\033[0m")
	case SegmentText:
		b.WriteString(content)
	case SegmentCode:
		if seg.FileName != "" {
			b.WriteString(fmt.Sprintf("\n┌─ %s\n", seg.FileName))
		}
		b.WriteString(content)
		if charLimit < 0 || charLimit >= len(seg.Content) {
			b.WriteString("\n└─\n")
		}
	}
}

// CurrentSegmentType returns the type of the active segment
func (s *Streamer) CurrentSegmentType() SegmentType {
	if s.segIdx < len(s.segments) {
		return s.segments[s.segIdx].Type
	}
	return SegmentText
}

// Displayed returns the total characters displayed so far
func (s *Streamer) Displayed() int {
	total := 0
	for i := 0; i < s.segIdx && i < len(s.segments); i++ {
		total += len(s.segments[i].Content)
	}
	total += s.pos
	return total
}
