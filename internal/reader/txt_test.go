package reader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTXTReaderLoadAndChapters(t *testing.T) {
	r := &TXTReader{}
	path := filepath.Join("..", "..", "testdata", "sample.txt")
	if err := r.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	chapters := r.Chapters()
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}
	if chapters[0].Title != "第一章 开始" {
		t.Errorf("chapter 0 title = %q, want %q", chapters[0].Title, "第一章 开始")
	}
	if chapters[1].Title != "第二章 发展" {
		t.Errorf("chapter 1 title = %q, want %q", chapters[1].Title, "第二章 发展")
	}
}

func TestTXTReaderPages(t *testing.T) {
	r := &TXTReader{}
	path := filepath.Join("..", "..", "testdata", "sample.txt")
	if err := r.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	totalPages := r.TotalPages(0, 80, 5)
	if totalPages < 1 {
		t.Errorf("TotalPages(0) = %d, want >= 1", totalPages)
	}

	page0 := r.ReadPage(0, 0, 80, 5)
	if page0 == "" {
		t.Error("ReadPage(0, 0) returned empty string")
	}
}

func TestTXTReaderInvalidPath(t *testing.T) {
	r := &TXTReader{}
	err := r.Load("nonexistent.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
