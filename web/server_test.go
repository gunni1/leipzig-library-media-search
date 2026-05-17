package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gunni1/leipzig-library-media-search/domain"
	"github.com/gunni1/leipzig-library-media-search/watchlist"
	. "github.com/stretchr/testify/assert"
)

func TestArrangeByBranch(t *testing.T) {

	t1_stadt := domain.Media{Title: "Terminator", Branch: "Stadt", IsAvailable: true}
	t2_stadt := domain.Media{Title: "Terminator 2", Branch: "Stadt", IsAvailable: false}
	t1_gohlis := domain.Media{Title: "Terminator", Branch: "Gohlis", IsAvailable: true}
	medias := []domain.Media{t1_stadt, t2_stadt, t1_gohlis}

	result := arrangeByBranch(medias)

	expected := []MediaByBranch{
		{Branch: "Stadt", Media: []domain.Media{t1_stadt, t2_stadt}},
		{Branch: "Gohlis", Media: []domain.Media{t1_gohlis}},
	}

	Equal(t, 2, len(result))
	ElementsMatch(t, result, expected)
}

func TestEncodeBranchName(t *testing.T) {
	Equal(t, 20, encodeBranch("Bibliothek Plagwitz"))
	Equal(t, 0, encodeBranch("Stadtbibliothek"))
	Equal(t, 41, encodeBranch("Bibliothek Gohlis"))
}

func TestAvailabilityHandler_returnsBadRequestWithoutTitle(t *testing.T) {
	store, _ := watchlist.NewFileStore(t.TempDir())
	mux := InitMux(store, "")
	req := httptest.NewRequest("GET", "/api/availability?type=movie", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAvailabilityHandler_returnsJSON(t *testing.T) {
	store, _ := watchlist.NewFileStore(t.TempDir())
	mux := InitMux(store, "")
	req := httptest.NewRequest("GET", "/api/availability?title=Dune&type=movie", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	// The library scrape will return empty for a test title — we just verify JSON shape.
	Equal(t, http.StatusOK, rec.Code)
	Equal(t, "application/json", rec.Header().Get("Content-Type"))
	var body map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &body)
	Nil(t, err)
	_, hasAvailable := body["available"]
	True(t, hasAvailable)
}

func TestSubscribeHandler_returnsServiceUnavailableWhenNotifierNotConfigured(t *testing.T) {
	store, _ := watchlist.NewFileStore(t.TempDir())
	mux := InitMux(store, "") // no notifier URL
	body := strings.NewReader("title=Dune&type=movie&email=test@example.com")
	req := httptest.NewRequest("POST", "/watchlist/subscribe", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestSubscribeHandler_returnsBadRequestWithoutEmail(t *testing.T) {
	store, _ := watchlist.NewFileStore(t.TempDir())
	// Use a fake notifier URL so the handler proceeds past the config check
	mux := InitMux(store, "http://notifier-does-not-exist")
	body := strings.NewReader("title=Dune&type=movie")
	req := httptest.NewRequest("POST", "/watchlist/subscribe", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	Equal(t, http.StatusBadRequest, rec.Code)
}
