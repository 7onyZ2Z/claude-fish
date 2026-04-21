package internal

type BossMode struct {
	streamer *Streamer
	active   bool
	hasCode  bool
}

func NewBossMode(segments []Segment, speed int) *BossMode {
	return &BossMode{
		streamer: NewStreamer(segments, speed),
		hasCode:  len(segments) > 0,
	}
}

func (b *BossMode) Active() bool        { return b.active }
func (b *BossMode) HasCode() bool       { return b.hasCode }
func (b *BossMode) Streamer() *Streamer { return b.streamer }

func (b *BossMode) Activate() {
	b.active = true
	if b.streamer.Done() {
		b.streamer.Reset()
	}
}

func (b *BossMode) Deactivate() {
	b.active = false
}
