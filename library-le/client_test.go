package libraryle

import (
	"net/http"
	"testing"

	. "github.com/stretchr/testify/assert"
)

const (
	jSessionId    string = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	userSessionId string = "2267N112S85e7645be446dd6c4e2e4bc558a206f3c4a88788"
)

func TestGameRequestHasSearchParameters(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	searchString := "Nintendo Switch"
	result := createGameSearchRequest(40, searchString, session)

	Equal(t, "submit", result.URL.Query().Get("methodToCall"))
	Equal(t, "submitSearch", result.URL.Query().Get("methodToCallParameter"))
	Equal(t, "902", result.URL.Query().Get("searchCategories[0]"))
	Equal(t, "Suchen", result.URL.Query().Get("submitSearch"))
	Equal(t, "searchPreferences", result.URL.Query().Get("callingPage"))
	Equal(t, "500", result.URL.Query().Get("numberOfHits"))
	Equal(t, "20", result.URL.Query().Get("timeOut"))

	Equal(t, userSessionId, result.URL.Query().Get("CSId"))
	Equal(t, searchString, result.URL.Query().Get("searchString[0]"))
	Equal(t, "40", result.URL.Query().Get("selectedSearchBranchlib"))
}

func TestGameRequestHasCookiesSet(t *testing.T) {
	session := webOpacSession{jSessionId: jSessionId, userSessionId: userSessionId}
	request := createGameSearchRequest(40, "Nintendo Switch", session)
	assertSessionCookiesExists(request, t)
}

func assertSessionCookiesExists(request *http.Request, t *testing.T) {
	Equal(t, 2, len(request.Cookies()))
	foundJSessionId := false
	foundUserSessionId := false
	for _, cookie := range request.Cookies() {
		switch cookie.Name {
		case "JSESSIONID":
			foundJSessionId = true
		case "USERSESSIONID":
			foundUserSessionId = true
		}
	}
	True(t, foundJSessionId)
	True(t, foundUserSessionId)
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
