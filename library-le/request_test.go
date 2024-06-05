package libraryle

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestMovieSearchRequestHasCookiesSet(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	request := NewMovieSearchRequest("Terminator", session)
	assertSessionCookiesExists(request, t)
}

func TestMovieSearchRequestHasQueryParamsSet(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	request := NewMovieSearchRequest("Terminator", session)
	Equal(t, "submit", request.URL.Query().Get("methodToCall"))
	Equal(t, "331", request.URL.Query().Get("searchCategories[0]"))
	Equal(t, "500", request.URL.Query().Get("numberOfHits"))
	Equal(t, "3", request.URL.Query().Get("searchRestrictionID[2]"))
	Equal(t, "29", request.URL.Query().Get("searchRestrictionValue1[2]"))
	Equal(t, "0", request.URL.Query().Get("selectedViewBranchlib"))
	Empty(t, request.URL.Query().Get("selectedSearchBranchlib"))
}
