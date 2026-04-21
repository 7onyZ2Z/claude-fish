package reader

// Chapter represents a single chapter in a novel.
type Chapter struct {
	Title string
	Index int
}

// Reader is the interface for novel format parsers.
type Reader interface {
	Load(path string) error
	Chapters() []Chapter
	ReadPage(chapter, page, width, linesPerPage int) string
	TotalPages(chapter, width, linesPerPage int) int
}
