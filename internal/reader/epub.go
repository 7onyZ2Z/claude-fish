package reader

import (
	"archive/zip"
	"encoding/xml"
	"io"
	"regexp"
	"strings"
)

type EPUBReader struct {
	chapters []txtChapter
}

// EPUB XML structures
type container struct {
	XMLName   xml.Name        `xml:"container"`
	RootFiles []containerRoot `xml:"rootfiles>rootfile"`
}

type containerRoot struct {
	FullPath string `xml:"full-path,attr"`
}

type opfPackage struct {
	XMLName  xml.Name     `xml:"package"`
	Manifest []opfItem    `xml:"manifest>item"`
	Spine    []opfItemRef `xml:"spine>itemref"`
}

type opfItem struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

type opfItemRef struct {
	IDRef string `xml:"idref,attr"`
}

func (r *EPUBReader) Load(path string) error {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer zr.Close()

	// 1. Find OPF from container.xml
	opfPath := r.findOPF(&zr.Reader)

	// 2. Parse OPF to get spine order
	pkg := r.parseOPF(&zr.Reader, opfPath)

	// 3. Build manifest map
	manifest := make(map[string]string) // id -> href
	for _, item := range pkg.Manifest {
		if item.MediaType == "application/xhtml+xml" {
			manifest[item.ID] = item.Href
		}
	}

	// 4. Follow spine, extract text from each chapter
	r.chapters = nil
	for _, ref := range pkg.Spine {
		href, ok := manifest[ref.IDRef]
		if !ok {
			continue
		}
		raw := r.extractRaw(&zr.Reader, href)
		content := r.stripHTML(raw)
		lines := strings.Split(content, "\n")
		var nonEmpty []string
		for _, l := range lines {
			if l = strings.TrimSpace(l); l != "" {
				nonEmpty = append(nonEmpty, l)
			}
		}
		title := extractTitle(raw)
		if title == "" {
			title = "Chapter"
		}
		r.chapters = append(r.chapters, txtChapter{
			title: title,
			lines: nonEmpty,
		})
	}

	return nil
}

func (r *EPUBReader) findOPF(zr *zip.Reader) string {
	for _, f := range zr.File {
		if f.Name == "META-INF/container.xml" {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			var c container
			xml.Unmarshal(data, &c)
			if len(c.RootFiles) > 0 {
				return c.RootFiles[0].FullPath
			}
		}
	}
	return "content.opf"
}

func (r *EPUBReader) parseOPF(zr *zip.Reader, opfPath string) *opfPackage {
	for _, f := range zr.File {
		if f.Name == opfPath {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			var pkg opfPackage
			xml.Unmarshal(data, &pkg)
			return &pkg
		}
	}
	return &opfPackage{}
}

var htmlTagRe = regexp.MustCompile(`<[^>]+>`)
var multiSpaceRe = regexp.MustCompile(`\s+`)

func (r *EPUBReader) extractRaw(zr *zip.Reader, href string) string {
	for _, f := range zr.File {
		if f.Name == href || strings.HasSuffix(f.Name, "/"+href) {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			defer rc.Close()
			data, _ := io.ReadAll(rc)
			return string(data)
		}
	}
	return ""
}

func (r *EPUBReader) extractText(zr *zip.Reader, href string) string {
	raw := r.extractRaw(zr, href)
	if raw == "" {
		return ""
	}
	return r.stripHTML(raw)
}

func (r *EPUBReader) stripHTML(raw string) string {
	text := raw
	text = regexp.MustCompile(`<(?:p|br|div|h[1-6])[^>]*>`).ReplaceAllString(text, "\n")
	text = htmlTagRe.ReplaceAllString(text, "")
	text = multiSpaceRe.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

var titleRe = regexp.MustCompile(`(?i)<title>([^<]+)</title>`)

func extractTitle(content string) string {
	matches := titleRe.FindStringSubmatch(content)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (r *EPUBReader) Chapters() []Chapter {
	result := make([]Chapter, len(r.chapters))
	for i, c := range r.chapters {
		result[i] = Chapter{Title: c.title, Index: i}
	}
	return result
}

func (r *EPUBReader) ReadPage(chapter, page, width, linesPerPage int) string {
	if chapter < 0 || chapter >= len(r.chapters) {
		return ""
	}
	lines := wrapLines(r.chapters[chapter].lines, width)
	start := page * linesPerPage
	end := start + linesPerPage
	if end > len(lines) {
		end = len(lines)
	}
	if start >= len(lines) {
		return ""
	}
	return strings.Join(lines[start:end], "\n")
}

func (r *EPUBReader) TotalPages(chapter, width, linesPerPage int) int {
	if chapter < 0 || chapter >= len(r.chapters) {
		return 0
	}
	lines := wrapLines(r.chapters[chapter].lines, width)
	return (len(lines) + linesPerPage - 1) / linesPerPage
}
