package libraryle

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/game-index-library-le/pkg/domain"
)

const resultItemSelector string = "td[style*='width:100%']"

// Takes a http.Response from a webopac search and
// try to parse it to an array of games that are listed as available.
func parseSearchResult(searchResult *http.Response) ([]domain.Game, error) {
	doc, docErr := goquery.NewDocumentFromReader(searchResult.Body)
	if docErr != nil {
		log.Fatal("Could not create document from response.")
		return nil, docErr
	}
	games := make([]domain.Game, 0)
	doc.Find("table.data").Each(func(i int, data *goquery.Selection) {
		data.Find(resultItemSelector).Each(func(i int, resultItem *goquery.Selection) {
			title := resultItem.Find("a[href^='/webOPACClient/singleHit']").Text()
			available := isAvailable(resultItem.Find("span").Text())
			games = append(games, domain.Game{Title: title})
			fmt.Printf("%s : %t\n", title, available)
		})

	})
	return games, nil
}

// Response is either:
// "Ein oder mehrere Exemplare dieses Titels sind in der aktuellen Zweigstelle ausleihbar." or
// "Der gewählte Titel ist in der aktuellen Zweigstelle entliehen."
func isAvailable(responseCode string) bool {
	return strings.Contains(responseCode, "ausleihbar")
}
