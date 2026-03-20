package watchlist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Item represents a single entry in the watchlist.
type Item struct {
	Title    string
	Platform string
	Type     string // "movie" or "game"
}

// FileStore is a file-backed, session-keyed watchlist store.
// Each session is persisted as a JSON file in the configured data directory.
type FileStore struct {
	mu      sync.Mutex
	dataDir string
}

// NewFileStore creates a FileStore rooted at dataDir, creating the directory if necessary.
func NewFileStore(dataDir string) (*FileStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &FileStore{dataDir: dataDir}, nil
}

func (fst *FileStore) filePath(sessionID string) string {
	return filepath.Join(fst.dataDir, sessionID+".json")
}

func (fst *FileStore) readItems(sessionID string) []Item {
	raw, err := os.ReadFile(fst.filePath(sessionID))
	if err != nil {
		// File not found or unreadable — treat as empty list.
		return []Item{}
	}
	var items []Item
	if err := json.Unmarshal(raw, &items); err != nil {
		return []Item{}
	}
	return items
}

func (fst *FileStore) writeItems(sessionID string, items []Item) error {
	raw, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return os.WriteFile(fst.filePath(sessionID), raw, 0644)
}

// GetAll returns a copy of all items for the given session.
func (fst *FileStore) GetAll(sessionID string) []Item {
	fst.mu.Lock()
	defer fst.mu.Unlock()
	return fst.readItems(sessionID)
}

// Toggle adds the item when absent, removes it when present.
// Returns true when the item is now in the watchlist.
func (fst *FileStore) Toggle(sessionID string, item Item) bool {
	fst.mu.Lock()
	defer fst.mu.Unlock()
	list := fst.readItems(sessionID)
	for idx, existing := range list {
		if existing.Title == item.Title && existing.Type == item.Type {
			list = append(list[:idx], list[idx+1:]...)
			fst.writeItems(sessionID, list)
			return false
		}
	}
	list = append(list, item)
	fst.writeItems(sessionID, list)
	return true
}

// Remove deletes a specific item from the session's watchlist.
func (fst *FileStore) Remove(sessionID, title, itemType string) {
	fst.mu.Lock()
	defer fst.mu.Unlock()
	list := fst.readItems(sessionID)
	for idx, existing := range list {
		if existing.Title == title && existing.Type == itemType {
			list = append(list[:idx], list[idx+1:]...)
			fst.writeItems(sessionID, list)
			return
		}
	}
}

// Clear removes all items for the given session by deleting its file.
func (fst *FileStore) Clear(sessionID string) {
	fst.mu.Lock()
	defer fst.mu.Unlock()
	os.Remove(fst.filePath(sessionID))
}
