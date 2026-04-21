package internal

type BossMode struct {
	streamer *Streamer
	active   bool
	hasCode  bool
}

func NewBossMode(code, fileName string, speed int) *BossMode {
	return &BossMode{
		streamer: NewStreamer(code, fileName, speed),
		hasCode:  code != "",
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
