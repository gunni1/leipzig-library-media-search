package notifier

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestStore_SaveAndGetAll(t *testing.T) {
	store, err := NewSubscriptionStore(t.TempDir())
	Nil(t, err)

	sub := Subscription{
		Email:    "test@example.com",
		Title:    "Dune",
		Type:     "movie",
		Platform: "",
	}
	saved, err := store.Save(sub)
	Nil(t, err)
	NotEmpty(t, saved.ID)

	all, err := store.GetAll()
	Nil(t, err)
	Equal(t, 1, len(all))
	Equal(t, "Dune", all[0].Title)
}

func TestStore_Delete(t *testing.T) {
	store, _ := NewSubscriptionStore(t.TempDir())
	sub, _ := store.Save(Subscription{Email: "a@b.com", Title: "Test", Type: "game", Platform: "switch"})

	err := store.Delete(sub.ID)
	Nil(t, err)

	all, _ := store.GetAll()
	Equal(t, 0, len(all))
}

func TestStore_DeleteNonExistentIsNoError(t *testing.T) {
	store, _ := NewSubscriptionStore(t.TempDir())
	err := store.Delete("does-not-exist")
	Nil(t, err)
}
