package libraryle

import (
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-game-stock-api/pkg/domain"
)

const (
	resultItemSelector string = "td[style*='width:100%']"
	resultDataSelector string = "table.data"
	gameTitleSelector  string = "a[href^='/webOPACClient/singleHit']"
)

// Takes a http.Response from a webopac search and
// try to parse it to an array of games that are listed as available.
func parseSearchResult(searchResult *http.Response) ([]domain.Game, error) {
	doc, docErr := goquery.NewDocumentFromReader(searchResult.Body)
	if docErr != nil {
		log.Fatal("Could not create document from response.")
		return nil, docErr
	}
	games := make([]domain.Game, 0)
	doc.Find(resultDataSelector).Each(func(i int, data *goquery.Selection) {
		data.Find(resultItemSelector).Each(func(i int, resultItem *goquery.Selection) {
			title := resultItem.Find(gameTitleSelector).Text()
			if isAvailable(resultItem.Find("span").Text()) {
				games = append(games, domain.Game{Title: title})
			}
		})
	})
	return games, nil
}

// See tests for all possible response codes
func isAvailable(responseCode string) bool {
	return strings.Contains(responseCode, "ausleihbar") || strings.Contains(responseCode, "heute zurückgebucht")
}
