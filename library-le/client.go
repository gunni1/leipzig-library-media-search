package libraryle

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
)

const (
	LIB_BASE_URL = "https://webopac.stadtbibliothek-leipzig.de"
)

// Deprecated?
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

// Deprecated?
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
	request := createGameSearchRequest(branchCode, platform, libClient.session)
	httpClient := http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		log.Println("error during search")
		return nil
	}
	defer response.Body.Close()

	games, parseResultErr := parseGameSearchResult(response.Body)
	if parseResultErr != nil {
		log.Fatalln(parseResultErr)
		return nil
	}
	return games
}

func (libClient Client) FindMovies(title string) []domain.Movie {
	sessionErr := libClient.openSession()
	if sessionErr != nil {
		fmt.Println(sessionErr)
		return nil
	}
	searchRequest := createMovieSearchRequest(title, libClient.session)
	httpClient := http.Client{}
	searchResponse, err := httpClient.Do(searchRequest)
	if err != nil {
		log.Println("error during search")
		return nil
	}
	//Titel und Links aus den Ergebnissen extrahieren
	resultTitles := parseMovieSearch(searchResponse.Body)
	log.Println(resultTitles)

	//Parallel Ergebnislinks folgen und Details über Zweigstelle und Verfpgbarkeit sammeln
	return nil
}

// Load all existing copys of a result title over all library branches
func (result searchResult) loadMovieCopies(libSession webOpacSession) []domain.Movie {
	//request := createRequest(libSession, result.resultUrl)

	//httpClient := http.Client{}
	//movieResponse, err := httpClient.Do(request)

	//TODO: für ein Result die verfügbaren exemplare laden und movie objekte erzeugen
	return nil
}

func (client *Client) openSession() error {
	resp, err := http.Get(LIB_BASE_URL + "/webOPACClient")
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
