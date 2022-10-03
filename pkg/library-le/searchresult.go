package libraryle

import (
	"fmt"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/game-index-library-le/pkg/domain"
)

const resultItemSelector string = "td[style*='width:100%']"

// Takes a http.Response from a webopac search and
// try to parse it to an array of games that are listed as available.
func parseSearchResult(searchResult *http.Response) ([]domain.Game, error) {
	doc, docErr := goquery.NewDocumentFromResponse(searchResult)
	if docErr != nil {
		log.Fatal("Could not create document from response.")
		return nil, docErr
	}

	doc.Find("table.data").Each(func(i int, data *goquery.Selection) {
		fmt.Println("debug: found data table")
		data.Find(resultItemSelector).Each(func(i int, resultItem *goquery.Selection) {
			title := resultItem.Find("a[href^='/webOPACClient/singleHit']").Text()
			available := resultItem.Find("span").Text()

			fmt.Printf("%s : %s\n", title, available)
		})

	})
	return nil, nil
}
