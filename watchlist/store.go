package watchlist

import "sync"

// Item represents a single entry in the watchlist.
type Item struct {
	Title    string
	Platform string
	Type     string // "movie" or "game"
}

// Store is an in-memory, session-keyed watchlist store.
type Store struct {
	mu   sync.RWMutex
	data map[string][]Item
}

func NewStore() *Store {
	return &Store{data: make(map[string][]Item)}
}

// GetAll returns a copy of all items for the given session.
func (s *Store) GetAll(sessionID string) []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.data[sessionID]
	result := make([]Item, len(src))
	copy(result, src)
	return result
}

// Toggle adds the item when absent, removes it when present.
// Returns true when the item is now in the watchlist.
func (s *Store) Toggle(sessionID string, item Item) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.data[sessionID]
	for i, existing := range list {
		if existing.Title == item.Title && existing.Type == item.Type {
			s.data[sessionID] = append(list[:i], list[i+1:]...)
			return false
		}
	}
	s.data[sessionID] = append(list, item)
	return true
}

// Remove deletes a specific item from the session's watchlist.
func (s *Store) Remove(sessionID, title, itemType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.data[sessionID]
	for i, existing := range list {
		if existing.Title == title && existing.Type == itemType {
			s.data[sessionID] = append(list[:i], list[i+1:]...)
			return
		}
	}
}

// Clear removes all items for the given session.
func (s *Store) Clear(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, sessionID)
}
