package notifier

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Subscription represents a single email notification subscription.
type Subscription struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Title     string    `json:"title"`
	Type      string    `json:"type"`
	Platform  string    `json:"platform"`
	CreatedAt time.Time `json:"createdAt"`
}

// SubscriptionStore is a file-backed store for subscriptions.
type SubscriptionStore struct {
	mu       sync.Mutex
	filePath string
}

// NewSubscriptionStore creates a SubscriptionStore backed by a JSON file in dataDir.
func NewSubscriptionStore(dataDir string) (*SubscriptionStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	return &SubscriptionStore{filePath: filepath.Join(dataDir, "subscriptions.json")}, nil
}

func (store *SubscriptionStore) read() ([]Subscription, error) {
	raw, err := os.ReadFile(store.filePath)
	if os.IsNotExist(err) {
		return []Subscription{}, nil
	}
	if err != nil {
		return nil, err
	}
	var subs []Subscription
	if err := json.Unmarshal(raw, &subs); err != nil {
		return nil, err
	}
	return subs, nil
}

func (store *SubscriptionStore) write(subs []Subscription) error {
	raw, err := json.Marshal(subs)
	if err != nil {
		return err
	}
	return os.WriteFile(store.filePath, raw, 0644)
}

// Save persists a new subscription, assigning it a UUID and CreatedAt timestamp.
func (store *SubscriptionStore) Save(sub Subscription) (Subscription, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	idBytes := make([]byte, 16)
	rand.Read(idBytes)
	sub.ID = hex.EncodeToString(idBytes)
	sub.CreatedAt = time.Now().UTC()
	subs, err := store.read()
	if err != nil {
		return Subscription{}, err
	}
	subs = append(subs, sub)
	return sub, store.write(subs)
}

// GetAll returns all subscriptions.
func (store *SubscriptionStore) GetAll() ([]Subscription, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.read()
}

// Delete removes the subscription with the given ID. No-op if not found.
func (store *SubscriptionStore) Delete(id string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	subs, err := store.read()
	if err != nil {
		return err
	}
	filtered := make([]Subscription, 0, len(subs))
	for _, sub := range subs {
		if sub.ID != id {
			filtered = append(filtered, sub)
		}
	}
	return store.write(filtered)
}
