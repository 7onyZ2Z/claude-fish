package internal

import "testing"

func TestBossModeInit(t *testing.T) {
	b := NewBossMode([]Segment{CodeSegment("main.go", "package main\n")}, 25)
	if b.Active() {
		t.Error("new BossMode should not be active")
	}
	if !b.HasCode() {
		t.Error("HasCode() should be true when segments loaded")
	}
}

func TestBossModeNoCode(t *testing.T) {
	b := NewBossMode(nil, 25)
	if !b.HasCode() {
		t.Error("HasCode() should be true — falls back to default content")
	}
}

func TestBossModeActivateDeactivate(t *testing.T) {
	b := NewBossMode([]Segment{TextSegment("hello world")}, 25)

	b.Activate()
	if !b.Active() {
		t.Error("should be active after Activate()")
	}
	if b.Streamer().Displayed() != 0 {
		t.Error("streamer should start at 0")
	}

	b.Streamer().Advance(5)
	if b.Streamer().Displayed() != 5 {
		t.Errorf("streamer should have advanced, got %d", b.Streamer().Displayed())
	}

	b.Deactivate()
	if b.Active() {
		t.Error("should not be active after Deactivate()")
	}

	b.Activate()
	if b.Streamer().Displayed() != 5 {
		t.Errorf("streamer should resume from 5, got %d", b.Streamer().Displayed())
	}

	b.Streamer().Advance(100)
	if !b.Streamer().Done() {
		t.Error("streamer should be done")
	}

	b.Deactivate()
	b.Activate()
	if b.Streamer().Displayed() != 0 {
		t.Error("after loop, streamer should start from 0")
	}
}
