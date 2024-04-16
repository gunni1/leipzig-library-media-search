package libraryle

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-game-stock-api/domain"
)

const (
	resultItemSelector   string = "h2[class^=recordtitle]"
	titleSelector        string = "a[href^='/webOPACClient/singleHit']"
	availabilitySelector string = "span[class^=textgruen]"
)

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

// Takes a html as reader from a webopac search and
// try to parse it to an array of games that are listed as available.
func parseGameSearchResult(searchResult io.Reader) ([]domain.Game, error) {
	doc, docErr := goquery.NewDocumentFromReader(searchResult)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil, docErr
	}
	games := make([]domain.Game, 0)
	doc.Find(resultItemSelector).Each(func(i int, resultItem *goquery.Selection) {
		title := resultItem.Find(titleSelector).Text()
		if isGameAvailable(resultItem.Parent()) {
			games = append(games, domain.Game{Title: title})
		}
	})
	return games, nil
}

func isGameAvailable(searchHitNode *goquery.Selection) bool {
	return len(searchHitNode.Find(availabilitySelector).Nodes) > 0
}
