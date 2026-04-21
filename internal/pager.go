package internal

import (
	"claude-fish/internal/reader"
)

// Pager manages chapter/page navigation over a Reader.
type Pager struct {
	r            reader.Reader
	width        int
	linesPerPage int
	currentCh    int
	currentPg    int
	chapters     []reader.Chapter
	totalPages   []int
}

func NewPager(r reader.Reader, width, linesPerPage int) *Pager {
	chapters := r.Chapters()
	p := &Pager{
		r:            r,
		width:        width,
		linesPerPage: linesPerPage,
		currentCh:    0,
		currentPg:    0,
		chapters:     chapters,
		totalPages:   make([]int, len(chapters)),
	}
	p.recalcAll()
	return p
}

func (p *Pager) recalcAll() {
	for i := range p.chapters {
		p.totalPages[i] = p.r.TotalPages(i, p.width, p.linesPerPage)
	}
	if p.currentCh >= 0 && p.currentCh < len(p.totalPages) {
		if p.currentPg >= p.totalPages[p.currentCh] && p.totalPages[p.currentCh] > 0 {
			p.currentPg = p.totalPages[p.currentCh] - 1
		}
	}
}

func (p *Pager) Chapters() []reader.Chapter { return p.chapters }
func (p *Pager) Chapter() int               { return p.currentCh }
func (p *Pager) Page() int                  { return p.currentPg }

func (p *Pager) TotalPages() int {
	if p.currentCh < 0 || p.currentCh >= len(p.totalPages) {
		return 0
	}
	return p.totalPages[p.currentCh]
}

func (p *Pager) CurrentContent() string {
	return p.r.ReadPage(p.currentCh, p.currentPg, p.width, p.linesPerPage)
}

func (p *Pager) CurrentTitle() string {
	if p.currentCh < 0 || p.currentCh >= len(p.chapters) {
		return ""
	}
	return p.chapters[p.currentCh].Title
}

func (p *Pager) NextPage() bool {
	if p.currentPg < p.totalPages[p.currentCh]-1 {
		p.currentPg++
		return true
	}
	if p.currentCh < len(p.chapters)-1 {
		p.currentCh++
		p.currentPg = 0
		return true
	}
	return false
}

func (p *Pager) PrevPage() bool {
	if p.currentPg > 0 {
		p.currentPg--
		return true
	}
	if p.currentCh > 0 {
		p.currentCh--
		p.currentPg = p.totalPages[p.currentCh] - 1
		if p.currentPg < 0 {
			p.currentPg = 0
		}
		return true
	}
	return false
}

func (p *Pager) Resize(width, linesPerPage int) {
	p.width = width
	p.linesPerPage = linesPerPage
	p.recalcAll()
}

func (p *Pager) SetThemeLines(linesPerPage int) {
	p.linesPerPage = linesPerPage
	p.recalcAll()
}

func (p *Pager) GoToChapter(ch int) {
	if ch >= 0 && ch < len(p.chapters) {
		p.currentCh = ch
		p.currentPg = 0
	}
}

func (p *Pager) TotalChapters() int { return len(p.chapters) }
