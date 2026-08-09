package notifier

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestNotifierServer_PostSubscription(t *testing.T) {
	store, _ := NewSubscriptionStore(t.TempDir())
	mux := NewNotifierMux(store, nil)

	form := url.Values{}
	form.Set("email", "user@example.com")
	form.Set("title", "Dune")
	form.Set("type", "movie")
	form.Set("platform", "")
	req := httptest.NewRequest("POST", "/subscriptions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	Equal(t, http.StatusCreated, rec.Code)

	all, _ := store.GetAll()
	Equal(t, 1, len(all))
	Equal(t, "Dune", all[0].Title)
}

func TestNotifierServer_PostSubscription_MissingEmail(t *testing.T) {
	store, _ := NewSubscriptionStore(t.TempDir())
	mux := NewNotifierMux(store, nil)

	form := url.Values{}
	form.Set("title", "Dune")
	form.Set("type", "movie")
	req := httptest.NewRequest("POST", "/subscriptions", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotifierServer_DeleteSubscription(t *testing.T) {
	store, _ := NewSubscriptionStore(t.TempDir())
	saved, _ := store.Save(Subscription{Email: "a@b.com", Title: "X", Type: "movie"})
	mux := NewNotifierMux(store, nil)

	req := httptest.NewRequest("DELETE", "/subscriptions/"+saved.ID, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	Equal(t, http.StatusNoContent, rec.Code)

	all, _ := store.GetAll()
	Equal(t, 0, len(all))
}
