package libraryle

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestParseMovieCopiesResult(t *testing.T) {
	testResponse := loadTestData("testdata/movie_copies_example.html")
	movies := parseMovieCopiesPage("Terminator - Genesis", testResponse)
	Equal(t, 6, len(movies))

	available := 0
	for _, movie := range movies {
		if movie.IsAvailable {
			available++
		}
	}
	Equal(t, 2, available)
}

func TestMovieSearchRequestHasCookiesSet(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	request := createMovieSearchRequest("Terminator", session)
	assertSessionCookiesExists(request, t)
}

func TestMovieSearchRequestHasQueryParamsSet(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	request := createMovieSearchRequest("Terminator", session)
	Equal(t, "submit", request.URL.Query().Get("methodToCall"))
	Equal(t, "331", request.URL.Query().Get("searchCategories[0]"))
	Equal(t, "500", request.URL.Query().Get("numberOfHits"))
	Equal(t, "3", request.URL.Query().Get("searchRestrictionID[2]"))
	Equal(t, "29", request.URL.Query().Get("searchRestrictionValue1[2]"))
	Equal(t, "0", request.URL.Query().Get("selectedViewBranchlib"))
	Empty(t, request.URL.Query().Get("selectedSearchBranchlib"))
}
