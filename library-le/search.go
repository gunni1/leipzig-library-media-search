package libraryle

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
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
		log.Println(err)
		return nil
	}
	resultTitles := extractTitles(searchResponse.Body)

	movies := make([]domain.Media, 0)
	//TODO: Parallel Ergbnislinks folgen und Details sammeln
	for _, resultTitle := range resultTitles {
		movies = append(movies, resultTitle.loadMediaCopies(libClient.session)...)
	}
	return movies
}

// Search for a specific game title in all library branches
func (libClient Client) FindGames(title string, platform string) []domain.Media {
	sessionErr := libClient.openSession()
	if sessionErr != nil {
		fmt.Println(sessionErr)
		return nil
	}
	searchRequest := NewGameSearchRequest(title, platform, libClient.session)
	httpClient := http.Client{}
	searchResponse, err := httpClient.Do(searchRequest)
	if err != nil {
		log.Println(err)
		return nil
	}
	resultTitles := extractTitles(searchResponse.Body)
	games := make([]domain.Media, 0)
	for _, resultTitle := range resultTitles {
		games = append(games, resultTitle.loadMediaCopies(libClient.session)...)
	}
	return games
}

// Load all existing copys of a result title over all library branches
func (result searchResult) loadMediaCopies(libSession webOpacSession) []domain.Media {
	request := createRequest(libSession, result.resultUrl)

	httpClient := http.Client{}
	mediaResponse, err := httpClient.Do(request)
	if err != nil {
		log.Println("error during search")
		return nil
	}
	return parseMediaCopiesPage(result.title, mediaResponse.Body)
}

// Go through the search overview page and create a result object for each title found.
// The result contain details of each copie availabile of the media.
func extractTitles(searchResponse io.Reader) []searchResult {
	doc, docErr := goquery.NewDocumentFromReader(searchResponse)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil
	}
	titles := make([]searchResult, 0)
	doc.Find(resultItemSelector).Each(func(i int, resultItem *goquery.Selection) {
		title := clearTitle(resultItem.Find(titleSelector).Text())
		resultUrl, _ := resultItem.Find(titleSelector).Attr("href")
		titles = append(titles, searchResult{title: title, resultUrl: resultUrl})
	})
	return titles
}

// the media copies page is a list of library branches which have the specific copy of a title
// it have information about the availability of the media
func parseMediaCopiesPage(title string, page io.Reader) []domain.Media {
	doc, docErr := goquery.NewDocumentFromReader(page)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil
	}
	movies := make([]domain.Media, 0)

	doc.Find(copiesSelector).Each(func(i int, copy *goquery.Selection) {
		branch := copy.Find("div.col-12.col-md-4.my-md-2 > b").Text()
		status := isMediaAvailable(copy)
		movies = append(movies, domain.Media{Title: title, Branch: removeBranchSuffix(branch), IsAvailable: status})
	})
	return movies
}

// Remove location detail suffix from branch name
func removeBranchSuffix(branchName string) string {
	return strings.TrimSpace(strings.Split(branchName, "/")[0])
}

// Remove additional media information from titles in square brackets
func clearTitle(title string) string {
	brackets := regexp.MustCompile(`\[.*\]`)
	return strings.TrimSpace(brackets.ReplaceAllString(title, ""))
}

func isMediaAvailable(copy *goquery.Selection) bool {
	rentalStateLink := copy.Find("div:nth-child(5) > div > a")
	//Link indicates a rented state (can reserve a copy)
	if rentalStateLink.Length() != 0 {
		return false
	}
	statusText := copy.Find("div:nth-child(5)").Text()
	return strings.Contains(statusText, "ausleihbar")
}
