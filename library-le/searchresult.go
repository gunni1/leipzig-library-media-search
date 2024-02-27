package libraryle

import (
	"io"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-game-stock-api/domain"
)

const (
	resultItemSelector   string = "h2[class^=recordtitle]"
	gameTitleSelector    string = "a[href^='/webOPACClient/singleHit']"
	availabilitySelector string = "span[class^=textgruen]"
)

// Takes a html as reader from a webopac search and
// try to parse it to an array of games that are listed as available.
func parseSearchResult(searchResult io.Reader) ([]domain.Game, error) {
	doc, docErr := goquery.NewDocumentFromReader(searchResult)
	if docErr != nil {
		log.Fatal("Could not create document from response.")
		return nil, docErr
	}
	games := make([]domain.Game, 0)
	doc.Find(resultItemSelector).Each(func(i int, resultItem *goquery.Selection) {
		title := resultItem.Find(gameTitleSelector).Text()
		if isAvailable(resultItem.Parent()) {
			games = append(games, domain.Game{Title: title})
		}
	})
	return games, nil
}

func isAvailable(searchHitNode *goquery.Selection) bool {
	return len(searchHitNode.Find(availabilitySelector).Nodes) > 0
}
