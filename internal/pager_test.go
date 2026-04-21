package internal

import (
	"path/filepath"
	"testing"

	"claude-fish/internal/reader"
)

func TestPagerNavigation(t *testing.T) {
	r := &reader.TXTReader{}
	path := filepath.Join("..", "testdata", "sample.txt")
	if err := r.Load(path); err != nil {
		t.Fatal(err)
	}

	p := NewPager(r, 80, 5)

	chapters := p.Chapters()
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}

	if p.Chapter() != 0 || p.Page() != 0 {
		t.Errorf("initial state = ch%d pg%d, want ch0 pg0", p.Chapter(), p.Page())
	}

	totalPages := p.TotalPages()
	for i := 0; i < totalPages; i++ {
		content := p.CurrentContent()
		if content == "" && i == 0 {
			t.Error("page 0 content is empty")
		}
		if i < totalPages-1 {
			if !p.NextPage() {
				t.Errorf("NextPage() returned false at page %d of %d", i, totalPages)
			}
		}
	}

	if totalPages > 1 {
		if !p.PrevPage() {
			t.Error("PrevPage() returned false")
		}
		if p.Page() != totalPages-2 {
			t.Errorf("after PrevPage, page = %d, want %d", p.Page(), totalPages-2)
		}
	}
}

func TestPagerResize(t *testing.T) {
	r := &reader.TXTReader{}
	path := filepath.Join("..", "testdata", "sample.txt")
	if err := r.Load(path); err != nil {
		t.Fatal(err)
	}

	p := NewPager(r, 80, 5)
	pagesBefore := p.TotalPages()

	p.Resize(40, 5)
	pagesAfter := p.TotalPages()

	if pagesAfter < pagesBefore {
		t.Errorf("narrower width should have >= pages, got %d < %d", pagesAfter, pagesBefore)
	}
}
