package reader

import (
	"path/filepath"
	"testing"
)

func TestMarkdownReaderLoadAndChapters(t *testing.T) {
	r := &MarkdownReader{}
	path := filepath.Join("..", "..", "testdata", "sample.md")
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
}

func TestMarkdownReaderPages(t *testing.T) {
	r := &MarkdownReader{}
	path := filepath.Join("..", "..", "testdata", "sample.md")
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
