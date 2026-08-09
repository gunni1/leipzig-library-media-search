package notifier

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestChecker_ReturnsTrueWhenAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equal(t, "/api/availability", r.URL.Path)
		Equal(t, "Dune", r.URL.Query().Get("title"))
		Equal(t, "movie", r.URL.Query().Get("type"))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"available": true,
			"branches":  []string{"Südvorstadt"},
		})
	}))
	defer server.Close()

	checker := NewAvailabilityChecker(server.URL)
	available, err := checker.IsAvailable(Subscription{Title: "Dune", Type: "movie"})
	Nil(t, err)
	True(t, available)
}

func TestChecker_ReturnsFalseWhenNotAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"available": false,
			"branches":  []string{},
		})
	}))
	defer server.Close()

	checker := NewAvailabilityChecker(server.URL)
	available, err := checker.IsAvailable(Subscription{Title: "Dune", Type: "movie"})
	Nil(t, err)
	False(t, available)
}

func TestChecker_ReturnsErrorOnHTTPFailure(t *testing.T) {
	checker := NewAvailabilityChecker("http://127.0.0.1:0") // nothing listening
	_, err := checker.IsAvailable(Subscription{Title: "X", Type: "movie"})
	NotNil(t, err)
}
