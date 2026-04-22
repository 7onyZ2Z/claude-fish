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
	totalPages   []int // -1 = not yet computed
	cacheValid   []bool
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
		cacheValid:   make([]bool, len(chapters)),
	}
	return p
}

func (p *Pager) getTotalPages(ch int) int {
	if ch < 0 || ch >= len(p.chapters) {
		return 0
	}
	if p.cacheValid[ch] {
		return p.totalPages[ch]
	}
	p.totalPages[ch] = p.r.TotalPages(ch, p.width, p.linesPerPage)
	p.cacheValid[ch] = true
	return p.totalPages[ch]
}

func (p *Pager) Chapters() []reader.Chapter { return p.chapters }
func (p *Pager) Chapter() int               { return p.currentCh }
func (p *Pager) Page() int                  { return p.currentPg }

func (p *Pager) TotalPages() int {
	return p.getTotalPages(p.currentCh)
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
	total := p.getTotalPages(p.currentCh)
	if p.currentPg < total-1 {
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
		total := p.getTotalPages(p.currentCh)
		p.currentPg = total - 1
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
	for i := range p.cacheValid {
		p.cacheValid[i] = false
	}
}

func (p *Pager) SetThemeLines(linesPerPage int) {
	p.linesPerPage = linesPerPage
	for i := range p.cacheValid {
		p.cacheValid[i] = false
	}
}

func (p *Pager) GoToChapter(ch int) {
	if ch >= 0 && ch < len(p.chapters) {
		p.currentCh = ch
		p.currentPg = 0
	}
}

func (p *Pager) SetPage(pg int) {
	total := p.getTotalPages(p.currentCh)
	if pg >= 0 && pg < total {
		p.currentPg = pg
	}
}

func (p *Pager) TotalChapters() int { return len(p.chapters) }
