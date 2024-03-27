package libraryle

import (
	"io"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-game-stock-api/domain"
)

const (
	resultItemSelector   string = "h2[class^=recordtitle]"
	titleSelector        string = "a[href^='/webOPACClient/singleHit']"
	availabilitySelector string = "span[class^=textgruen]"
	copiesSelector       string = "#tab-content > div > div:nth-child(n+2)"
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
		if isGameAvailable(resultItem.Parent()) {
			games = append(games, domain.Game{Title: title})
		}
	})
	return games, nil
}

func isGameAvailable(searchHitNode *goquery.Selection) bool {
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

func parseMovieCopiesPage(title string, page io.Reader) []domain.Movie {
	doc, docErr := goquery.NewDocumentFromReader(page)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil
	}
	movies := make([]domain.Movie, 0)

	doc.Find(copiesSelector).Each(func(i int, copy *goquery.Selection) {
		branch := copy.Find("div.col-12.col-md-4.my-md-2 > b").Text()
		status := isMovieAvailable(copy)
		movies = append(movies, domain.Movie{Title: title, Branch: branch, IsAvailable: status})
	})

	return movies
}

func isMovieAvailable(copy *goquery.Selection) bool {
	rentalStateLink := copy.Find("div:nth-child(5) > div > a")
	//Link indicates a rented state (can reserve a copy)
	if rentalStateLink.Length() != 0 {
		return false
	}
	statusText := copy.Find("div:nth-child(5)").Text()
	return strings.Contains(statusText, "ausleihbar")
}
