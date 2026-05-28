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

	added := store.Toggle("2502d7238f5963949085e14faa2980bd", item)
	True(t, added, "first toggle should add item")

	items := store.GetAll("2502d7238f5963949085e14faa2980bd")
	Equal(t, 1, len(items))
	Equal(t, item, items[0])

	removed := store.Toggle("2502d7238f5963949085e14faa2980bd", item)
	False(t, removed, "second toggle should remove item")

	items = store.GetAll("2502d7238f5963949085e14faa2980bd")
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
	store.Toggle("2502d7238f5963949085e14faa29aaaa", Item{Title: "Matrix", Type: "movie"})
	store.Toggle("2502d7238f5963949085e14faa29aaaa", Item{Title: "Dune", Type: "movie"})

	store.Clear("2502d7238f5963949085e14faa29aaaa")

	items := store.GetAll("2502d7238f5963949085e14faa29aaaa")
	Equal(t, 0, len(items))
}

func TestFileStorePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	item := Item{Title: "Dark", Type: "movie"}

	// Write via first instance.
	store1, err := NewFileStore(tmpDir)
	Nil(t, err)
	store1.Toggle("2502d7238f5963949085e14faa29bbbb", item)

	// Read back via a second instance pointing at the same directory.
	store2, err := NewFileStore(tmpDir)
	Nil(t, err)
	items := store2.GetAll("2502d7238f5963949085e14faa29bbbb")

	Equal(t, 1, len(items))
	Equal(t, item, items[0])
}

func TestFileStoreIsolatesSessions(t *testing.T) {
	store := newTestFileStore(t)
	store.Toggle("2502d7238f5963949085e14faa29aaaa", Item{Title: "Film A", Type: "movie"})
	store.Toggle("2502d7238f5963949085e14faa29bbbb", Item{Title: "Film B", Type: "movie"})

	itemsA := store.GetAll("2502d7238f5963949085e14faa29aaaa")
	itemsB := store.GetAll("2502d7238f5963949085e14faa29bbbb")

	Equal(t, 1, len(itemsA))
	Equal(t, "Film A", itemsA[0].Title)
	Equal(t, 1, len(itemsB))
	Equal(t, "Film B", itemsB[0].Title)
}
