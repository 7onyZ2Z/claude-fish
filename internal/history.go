package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ReadingPosition struct {
	Chapter   int       `json:"chapter"`
	Page      int       `json:"page"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HistoryStore struct {
	mu       sync.Mutex
	path     string
	Entries  map[string]ReadingPosition `json:"entries"`
}

func defaultHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(home, ".claude-fish")
	_ = os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "history.json")
}

func LoadHistory() *HistoryStore {
	path := defaultHistoryPath()
	h := &HistoryStore{path: path, Entries: make(map[string]ReadingPosition)}
	data, err := os.ReadFile(path)
	if err != nil {
		return h
	}
	_ = json.Unmarshal(data, &h.Entries)
	if h.Entries == nil {
		h.Entries = make(map[string]ReadingPosition)
	}
	return h
}

func (h *HistoryStore) Get(filePath string) (ReadingPosition, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	pos, ok := h.Entries[filePath]
	return pos, ok
}

func (h *HistoryStore) Save(filePath string, chapter, page int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Entries[filePath] = ReadingPosition{
		Chapter:   chapter,
		Page:      page,
		UpdatedAt: time.Now(),
	}
	h.flush()
}

func (h *HistoryStore) flush() {
	if h.path == "" {
		return
	}
	data, _ := json.MarshalIndent(h.Entries, "", "  ")
	_ = os.WriteFile(h.path, data, 0644)
}
