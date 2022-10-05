package libraryle

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gunni1/game-index-library-le/pkg/domain"
)

type Client struct {
	session webOpacSession
}

type webOpacSession struct {
	jSessionId    string
	userSessionId string
}

func (libClient Client) FindAvailabelGames(branchCode int, console string) []domain.Game {
	sessionErr := libClient.openSession()
	if sessionErr != nil {
		fmt.Println(sessionErr)
		return nil
	}

	request := createSearchRequest(branchCode, console, libClient.session.jSessionId, libClient.session.userSessionId)
	httpClient := http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		log.Fatal("error during search")
		return nil
	}
	defer response.Body.Close()

	games, parseResultErr := parseSearchResult(response)
	if parseResultErr != nil {
		log.Fatalln(parseResultErr)
		return nil
	}
	return games
}

func createSearchRequest(branchCode int, searchString string, jSessionId string, userSessionId string) *http.Request {
	request, _ := http.NewRequest("GET", "https://webopac.stadtbibliothek-leipzig.de/webOPACClient/search.do", nil)
	jSessionCookie := &http.Cookie{
		Name:  "JSESSIONID",
		Value: jSessionId,
	}
	userSessionCookie := &http.Cookie{
		Name:  "USERSESSIONID",
		Value: userSessionId,
	}
	request.AddCookie(jSessionCookie)
	request.AddCookie(userSessionCookie)

	query := request.URL.Query()
	//Fix Query Params to make the search working
	query.Add("methodToCall", "submit")
	query.Add("methodToCallParameter", "submitSearch")
	query.Add("searchCategories[0]", "902")
	query.Add("submitSearch", "Suchen")
	query.Add("callingPage", "searchPreferences")
	query.Add("numberOfHits", "500")
	query.Add("timeOut", "20")
	//Query Params dependend on user input / session
	query.Add("CSId", userSessionId)
	query.Add("searchString[0]", searchString)
	query.Add("selectedSearchBranchlib", strconv.FormatInt(int64(branchCode), 10))
	request.URL.RawQuery = query.Encode()
	return request
}

func (client *Client) openSession() error {
	resp, err := http.Get("https://webopac.stadtbibliothek-leipzig.de/webOPACClient")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var session = webOpacSession{}
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case "JSESSIONID":
			session.jSessionId = cookie.Value
		case "USERSESSIONID":
			session.userSessionId = cookie.Value
		}
	}
	client.session = session
	if client.session.jSessionId == "" || client.session.userSessionId == "" {
		return errors.New("did not receive valid session ids via cookie")
	}
	return nil
}
