package libraryle

import (
	"io"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-game-stock-api/domain"
)

const (
	resultItemSelector   string = "h2[class^=recordtitle]"
	titleSelector        string = "a[href^='/webOPACClient/singleHit']"
	availabilitySelector string = "span[class^=textgruen]"
	copiesSelector       string = "div[id='tab-content']/div/div[position()>1]"
)

type searchResult struct {
	title     string
	resultUrl string
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
		if isAvailable(resultItem.Parent()) {
			games = append(games, domain.Game{Title: title})
		}
	})
	return games, nil
}

func isAvailable(searchHitNode *goquery.Selection) bool {
	return len(searchHitNode.Find(availabilitySelector).Nodes) > 0
}

func parseMovieSearch(searchResponse io.Reader) []searchResult {
	doc, docErr := goquery.NewDocumentFromReader(searchResponse)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil
	}
	titles := make([]searchResult, 0)
	doc.Find(resultItemSelector).Each(func(i int, resultItem *goquery.Selection) {
		title := resultItem.Find(titleSelector).Text()
		resultUrl, _ := resultItem.Find(titleSelector).Attr("href")
		titles = append(titles, searchResult{title: title, resultUrl: resultUrl})
	})
	return titles
}

func parseMovieCopiesPage(page io.Reader) []domain.Movie {
	doc, docErr := goquery.NewDocumentFromReader(page)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil
	}
	//movies := make([]domain.Movie, 0)
	doc.Find(copiesSelector).Each(func(i int, copy *goquery.Selection) {
		//XPath für jedes relevante Attribut definieren
		mediaNum := copy.Get(1)
		branch := copy.Get(2)
		status := copy.Get(4)
		log.Printf("%s %s %s\n", mediaNum.Data, branch.Data, status.Data)

	})

	return nil
}
