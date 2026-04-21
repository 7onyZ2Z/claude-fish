package reader

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestEPUB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	epubPath := filepath.Join(dir, "test.epub")

	f, err := os.Create(epubPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)

	mw, _ := w.Create("mimetype")
	mw.Write([]byte("application/epub+zip"))

	cw, _ := w.Create("META-INF/container.xml")
	cw.Write([]byte(`<?xml version="1.0"?>
<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container" version="1.0">
  <rootfiles>
    <rootfile full-path="content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`))

	ow, _ := w.Create("content.opf")
	ow.Write([]byte(`<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="uid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
  </metadata>
  <manifest>
    <item id="ch1" href="ch1.xhtml" media-type="application/xhtml+xml"/>
    <item id="ch2" href="ch2.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
    <itemref idref="ch2"/>
  </spine>
</package>`))

	c1w, _ := w.Create("ch1.xhtml")
	c1w.Write([]byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter 1</title></head>
<body><p>这是第一章的内容。讲述了一个测试故事。</p><p>第二段内容在这里。</p></body>
</html>`))

	c2w, _ := w.Create("ch2.xhtml")
	c2w.Write([]byte(`<?xml version="1.0"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Chapter 2</title></head>
<body><p>这是第二章的内容。故事继续发展。</p></body>
</html>`))

	w.Close()
	return epubPath
}

func TestEPUBReaderLoadAndChapters(t *testing.T) {
	r := &EPUBReader{}
	path := createTestEPUB(t)
	if err := r.Load(path); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	chapters := r.Chapters()
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}
	if !strings.Contains(chapters[0].Title, "1") {
		t.Errorf("chapter 0 title = %q, want to contain '1'", chapters[0].Title)
	}
}

func TestEPUBReaderPages(t *testing.T) {
	r := &EPUBReader{}
	path := createTestEPUB(t)
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
