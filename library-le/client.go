package libraryle

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
)

var BranchCodes = map[int]string{
	0:  "Stadtbibliothek",
	20: "Bibliothek Plagwitz",
	21: "Bibliothek Wiederitzsch",
	22: "Bibliothek Böhlitz-Ehrenberg",
	23: "Bibl. Lützschena-Stahmeln",
	25: "Bibliothek Holzhausen",
	30: "Bibliothek Südvorstadt",
	41: "Bibliothek Gohlis",
	50: "Bibliothek Volkmarsdorf",
	51: "Bibliothek Schönefeld",
	60: "Bibliothek Paunsdorf",
	61: "Bibliothek Reudnitz",
	70: "Bibliothek Mockau",
	82: "Bibliothek Grünau-Mitte",
	83: "Bibliothek Grünau-Nord",
	84: "Bibliothek Grünau-Süd",
	90: "Fahrbibliothek",
}

func BranchCodeKeys() []int {
	keys := make([]int, 0, len(BranchCodes))
	for key := range BranchCodes {
		keys = append(keys, key)
	}
	return keys
}

type Client struct {
	session webOpacSession
}

type webOpacSession struct {
	jSessionId    string
	userSessionId string
}

func (libClient Client) FindAvailabelGames(branchCode int, platform string) []domain.Game {
	sessionErr := libClient.openSession()
	if sessionErr != nil {
		fmt.Println(sessionErr)
		return nil
	}
	request := createSearchRequest(branchCode, platform, libClient.session.jSessionId, libClient.session.userSessionId)
	httpClient := http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		log.Fatal("error during search")
		return nil
	}
	defer response.Body.Close()

	games, parseResultErr := parseSearchResult(response.Body)
	//Add branchCode to games?
	if parseResultErr != nil {
		log.Fatalln(parseResultErr)
		return nil
	}
	return games
}

func (libClient Client) GetAllAvailableGamesPlatform(platform string) []domain.Game {
	searchResults := make(chan domain.Game)

	wg := &sync.WaitGroup{}
	for _, code := range BranchCodeKeys() {
		wg.Add(1)
		go getAvailableGames(code, platform, searchResults, wg, libClient)
	}
	go func() {
		wg.Wait()
		close(searchResults)
	}()
	games := make([]domain.Game, 0)
	for game := range searchResults {
		games = append(games, game)
	}
	return games
}

func getAvailableGames(branchCode int, platform string, results chan domain.Game, wg *sync.WaitGroup, client Client) {
	defer wg.Done()
	games := client.FindAvailabelGames(branchCode, platform)
	for _, game := range games {
		results <- game
	}
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
	query.Add("selectedViewBranchlib", strconv.FormatInt(int64(branchCode), 10))
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
