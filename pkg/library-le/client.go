package libraryle

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gunni1/game-index-library-le/pkg/domain"
)

type Client struct {
	baseUrl string
	session webOpacSession
}

type webOpacSession struct {
	jSessionId    string
	userSessionId string
}

func (client Client) FindAvailabelGames(branchCode int, console string) []domain.Game {
	sessionErr := client.openSession()
	if sessionErr != nil {
		fmt.Println(sessionErr)
		return nil
	}

	http.Get("https://webopac.stadtbibliothek-leipzig.de/webOPACClient/search.do?selectedViewBranchlib=41&selectedSearchBranchlib=41")
	request, _ := http.NewRequest("GET", "https://webopac.stadtbibliothek-leipzig.de/webOPACClient/search.do", nil)
	query := request.URL.Query()
	//Fix Query Params to make the search working
	query.Add("methodToCall", "submit")
	query.Add("methodToCallParameter", "submitSearch")
	query.Add("searchCategories%5B0%5D", "902")
	query.Add("submitSearch", "Suchen")
	query.Add("callingPage", "searchPreferences")
	query.Add("numberOfHits", "500")
	query.Add("timeOut", "20")
	//Query Params dependend on user input / session
	query.Add("CSId", client.session.userSessionId)
	query.Add("searchString%5B0%5D", console)
	query.Add("selectedSearchBranchlib", strconv.FormatInt(int64(branchCode), 10))
	fmt.Println(query.Encode())

	return nil
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
