package libraryle

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestMovieSearchRequestHasCookiesSet(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	request := NewMovieSearchRequest("Terminator", 0, session)
	assertSessionCookiesExists(request, t)
}

func TestMovieSearchRequestHasQueryParamsSet(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	request := NewMovieSearchRequest("Terminator", 0, session)
	Equal(t, "submit", request.URL.Query().Get("methodToCall"))
	Equal(t, "331", request.URL.Query().Get("searchCategories[0]"))
	Equal(t, "500", request.URL.Query().Get("numberOfHits"))
	Equal(t, "3", request.URL.Query().Get("searchRestrictionID[2]"))
	Equal(t, "29", request.URL.Query().Get("searchRestrictionValue1[2]"))
	Equal(t, "0", request.URL.Query().Get("selectedViewBranchlib"))
	Empty(t, request.URL.Query().Get("selectedSearchBranchlib"))
}

func TestMovieReturnDateRequest(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	request := NewReturnDateRequest("Terminator", "dvd", 41, session)
	//Expect results to be restricted to dvd/bluray
	Equal(t, "29", request.URL.Query().Get("searchRestrictionValue1[2]"))
	Equal(t, "dvd", request.URL.Query().Get("searchString[1]"))
	Equal(t, "800", request.URL.Query().Get("searchCategories[1]"))
}

func TestGameReturnDateRequest(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	request := NewReturnDateRequest("Mario", "switch", 41, session)
	//Expect results to be restricted to games
	Equal(t, "27", request.URL.Query().Get("searchRestrictionValue1[2]"))
	Equal(t, "switch", request.URL.Query().Get("searchString[1]"))
	Equal(t, "902", request.URL.Query().Get("searchCategories[1]"))
}
