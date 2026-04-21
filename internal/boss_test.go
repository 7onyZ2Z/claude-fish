package internal

import "testing"

func TestBossModeInit(t *testing.T) {
	b := NewBossMode("package main\n", "main.go", 25)
	if b.Active() {
		t.Error("new BossMode should not be active")
	}
	if !b.HasCode() {
		t.Error("HasCode() should be true when code is loaded")
	}
}

func TestBossModeNoCode(t *testing.T) {
	b := NewBossMode("", "", 25)
	if b.HasCode() {
		t.Error("HasCode() should be false when no code")
	}
}

func TestBossModeActivateDeactivate(t *testing.T) {
	b := NewBossMode("code", "test.go", 25)

	b.Activate()
	if !b.Active() {
		t.Error("should be active after Activate()")
	}
	if b.Streamer().Displayed() != 0 {
		t.Error("streamer should start at 0")
	}

	b.Streamer().Advance(2)
	if b.Streamer().Displayed() != 2 {
		t.Errorf("streamer should have advanced, got %d", b.Streamer().Displayed())
	}

	b.Deactivate()
	if b.Active() {
		t.Error("should not be active after Deactivate()")
	}

	b.Activate()
	if b.Streamer().Displayed() != 2 {
		t.Errorf("streamer should resume from 2, got %d", b.Streamer().Displayed())
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
