package libraryle

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-game-stock-api/domain"
)

const (
	copiesSelector string = "#tab-content > div > div:nth-child(n+2)"
)

type searchResult struct {
	title     string
	resultUrl string
}

// Search for a specific movie title in all library branches
func (libClient Client) FindMovies(title string) []domain.Media {
	sessionErr := libClient.openSession()
	if sessionErr != nil {
		fmt.Println(sessionErr)
		return nil
	}
	searchRequest := NewMovieSearchRequest(title, libClient.session)
	httpClient := http.Client{}
	searchResponse, err := httpClient.Do(searchRequest)
	if err != nil {
		log.Println("error during search")
		return nil
	}
	resultTitles := parseMediaSearch(searchResponse.Body)

	movies := make([]domain.Media, 0)
	for _, resultTitle := range resultTitles {
		movies = append(movies, resultTitle.loadMovieCopies(libClient.session)...)
	}
	//Parallel Ergebnislinks folgen und Details über Zweigstelle und Verfpgbarkeit sammeln
	return movies
}

// Load all existing copys of a result title over all library branches
func (result searchResult) loadMovieCopies(libSession webOpacSession) []domain.Media {
	request := createRequest(libSession, result.resultUrl)

	httpClient := http.Client{}
	movieResponse, err := httpClient.Do(request)
	if err != nil {
		log.Println("error during search")
		return nil
	}
	return parseMediaCopiesPage(result.title, movieResponse.Body)
}

func parseMediaSearch(searchResponse io.Reader) []searchResult {
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

func parseMediaCopiesPage(title string, page io.Reader) []domain.Media {
	doc, docErr := goquery.NewDocumentFromReader(page)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil
	}
	movies := make([]domain.Media, 0)

	doc.Find(copiesSelector).Each(func(i int, copy *goquery.Selection) {
		branch := copy.Find("div.col-12.col-md-4.my-md-2 > b").Text()
		status := isMovieAvailable(copy)
		movies = append(movies, domain.Media{Title: title, Branch: branch, IsAvailable: status})
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
