package libraryle

import (
	"net/http"
	"strconv"
)

func createRequest(libSession webOpacSession, path string) *http.Request {
	request, _ := http.NewRequest("GET", LIB_BASE_URL+path, nil)
	jSessionCookie := &http.Cookie{
		Name:  "JSESSIONID",
		Value: libSession.jSessionId,
	}
	userSessionCookie := &http.Cookie{
		Name:  "USERSESSIONID",
		Value: libSession.userSessionId,
	}
	request.AddCookie(jSessionCookie)
	request.AddCookie(userSessionCookie)
	return request
}

func createMovieSearchRequest(searchString string, libSession webOpacSession) *http.Request {
	request := createRequest(libSession, "/webOPACClient/search.do")
	request.URL.RawQuery = createMovieSearchQuery(*request, searchString, libSession.userSessionId)
	return request
}

func createGameSearchRequest(branchCode int, platform string, libSession webOpacSession) *http.Request {
	request := createRequest(libSession, "/webOPACClient/search.do")
	request.URL.RawQuery = createGameIndexQuery(*request, platform, libSession.userSessionId, branchCode)
	return request
}

func createMovieSearchQuery(request http.Request, searchString string, userSessionId string) string {
	query := request.URL.Query()
	query.Add("methodToCall", "submit")
	query.Add("methodToCallParameter", "submitSearch")

	query.Add("submitSearch", "Suchen")
	query.Add("callingPage", "searchPreferences")
	query.Add("numberOfHits", "500")
	query.Add("timeOut", "20")
	query.Add("CSId", userSessionId)
	query.Add("searchString[0]", searchString)
	query.Add("selectedViewBranchlib", strconv.FormatInt(int64(0), 10))
	//Search for category title
	query.Add("searchCategories[0]", "331")
	//Restrict search to dvd/bluray
	query.Add("searchRestrictionID[2]", "3")
	query.Add("searchRestrictionValue1[2]", "29")
	return query.Encode()
}

func createGameIndexQuery(request http.Request, platform string, userSessionId string, branchCode int) string {
	query := request.URL.Query()
	query.Add("methodToCall", "submit")
	query.Add("methodToCallParameter", "submitSearch")
	query.Add("submitSearch", "Suchen")
	query.Add("callingPage", "searchPreferences")
	query.Add("numberOfHits", "500")
	query.Add("timeOut", "20")
	query.Add("CSId", userSessionId)
	query.Add("selectedSearchBranchlib", strconv.FormatInt(int64(branchCode), 10))
	query.Add("selectedViewBranchlib", strconv.FormatInt(int64(branchCode), 10))
	//Search the platform as a catchword
	query.Add("searchString[0]", platform)
	query.Add("searchCategories[0]", "902")

	return query.Encode()
}
