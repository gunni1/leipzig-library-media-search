package watchlist

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func newTestFileStore(t *testing.T) *FileStore {
	t.Helper()
	fstore, err := NewFileStore(t.TempDir())
	Nil(t, err)
	return fstore
}

func TestFileStoreToggle(t *testing.T) {
	store := newTestFileStore(t)
	item := Item{Title: "Inception", Type: "movie"}

	added := store.Toggle("session1", item)
	True(t, added, "first toggle should add item")

	items := store.GetAll("session1")
	Equal(t, 1, len(items))
	Equal(t, item, items[0])

	removed := store.Toggle("session1", item)
	False(t, removed, "second toggle should remove item")

	items = store.GetAll("session1")
	Equal(t, 0, len(items))
}

func TestFileStoreRemove(t *testing.T) {
	store := newTestFileStore(t)
	item := Item{Title: "The Witcher 3", Type: "game", Platform: "PC"}

	store.Toggle("session2", item)
	store.Remove("session2", item.Title, item.Type)

	items := store.GetAll("session2")
	Equal(t, 0, len(items))
}

func TestFileStoreClear(t *testing.T) {
	store := newTestFileStore(t)
	store.Toggle("session3", Item{Title: "Matrix", Type: "movie"})
	store.Toggle("session3", Item{Title: "Dune", Type: "movie"})

	store.Clear("session3")

	items := store.GetAll("session3")
	Equal(t, 0, len(items))
}

func TestFileStorePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	item := Item{Title: "Dark", Type: "movie"}

	// Write via first instance.
	store1, err := NewFileStore(tmpDir)
	Nil(t, err)
	store1.Toggle("sessionX", item)

	// Read back via a second instance pointing at the same directory.
	store2, err := NewFileStore(tmpDir)
	Nil(t, err)
	items := store2.GetAll("sessionX")

	Equal(t, 1, len(items))
	Equal(t, item, items[0])
}

func TestFileStoreIsolatesSessions(t *testing.T) {
	store := newTestFileStore(t)
	store.Toggle("sessionA", Item{Title: "Film A", Type: "movie"})
	store.Toggle("sessionB", Item{Title: "Film B", Type: "movie"})

	itemsA := store.GetAll("sessionA")
	itemsB := store.GetAll("sessionB")

	Equal(t, 1, len(itemsA))
	Equal(t, "Film A", itemsA[0].Title)
	Equal(t, 1, len(itemsB))
	Equal(t, "Film B", itemsB[0].Title)
}
