package libraryle

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-media-search/domain"
	"github.com/pkg/errors"
)

// Search for a specific movie title in all library branches
func (libClient Client) FindMovies(title string) []domain.Media {
	sessionErr := libClient.newSession()
	if sessionErr != nil {
		fmt.Println(sessionErr)
		return nil
	}
	searchRequest := NewMovieSearchRequest(title, 0, libClient.session)
	searchResponse, err := httpClient.Do(searchRequest)
	if err != nil {
		log.Println(err)
		return nil
	}
	doc, docErr := goquery.NewDocumentFromReader(searchResponse.Body)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil
	}
	if isSingleResultPage(doc) {
		return parseMediaCopiesPage(title, doc)
	}
	resultTitles := extractTitles(doc)
	movies := make([]domain.Media, 0)
	//TODO: Parallel Ergbnislinks folgen und Details sammeln
	for _, resultTitle := range resultTitles {
		movies = append(movies, resultTitle.loadMediaCopies(libClient.session)...)
	}
	return movies
}

// Search for a specific game title in all library branches
func (libClient Client) FindGames(title string, platform string) []domain.Media {
	if sessionErr := libClient.newSession(); sessionErr != nil {
		fmt.Println(sessionErr)
		return nil
	}
	searchRequest := NewGameSearchRequest(title, platform, 0, libClient.session)
	searchResponse, err := httpClient.Do(searchRequest)
	if err != nil {
		log.Println(err)
		return nil
	}
	doc, docErr := goquery.NewDocumentFromReader(searchResponse.Body)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil
	}
	if isSingleResultPage(doc) {
		return parseMediaCopiesPage(title, doc)
	}
	resultTitles := extractTitles(doc)
	games := make([]domain.Media, 0)
	for _, resultTitle := range resultTitles {
		games = append(games, resultTitle.loadMediaCopies(libClient.session)...)
	}
	return games
}

func (libClient Client) RetrieveReturnDate(branchCode int, platform string, title string) (string, error) {
	request := NewReturnDateRequest(title, platform, branchCode, libClient.session)
	searchResponse, err := httpClient.Do(request)
	if err != nil {
		log.Printf("Error during search: %s", err.Error())
		return "-", err
	}
	doc, docErr := goquery.NewDocumentFromReader(searchResponse.Body)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return "", docErr
	}

	if isSingleResultPage(doc) {
		return findReturnDateInCopiesPage(doc)
	} else {
		resultTitles := extractTitles(doc)
		exactMatchTitles := filterExactTitle(title, resultTitles)
		return loadMediaReturnDate(exactMatchTitles, libClient.session)
	}
}

// Load all existing copys of a result title over all library branches
func (result searchResult) loadMediaCopies(libSession webOpacSession) []domain.Media {
	request := createRequest(libSession, result.resultUrl)

	mediaResponse, err := httpClient.Do(request)
	if err != nil {
		log.Printf("Error during search: %s", err.Error())
		return nil
	}
	doc, docErr := goquery.NewDocumentFromReader(mediaResponse.Body)
	if docErr != nil {
		log.Println("Could not create document from response.")
		return nil
	}
	return parseMediaCopiesPage(result.title, doc)
}

// load the return date for a searched title. Return the date of the first copy found.
func (result searchResult) loadReturnDate(libSession webOpacSession) (string, error) {
	request := createRequest(libSession, result.resultUrl)
	mediaResponse, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("http request failed for %s: %w", result.title, err)
	}
	doc, docErr := goquery.NewDocumentFromReader(mediaResponse.Body)
	if docErr != nil {
		return "", fmt.Errorf("could not parse response for %s: %w", result.title, docErr)
	}
	return findReturnDateInCopiesPage(doc)
}

func loadMediaReturnDate(titles []searchResult, libSession webOpacSession) (string, error) {
	//do a request for every searchresult
	//TODO: find earliest date
	for _, title := range titles {
		returnDate, err := title.loadReturnDate(libSession)
		if err == nil {
			return returnDate, nil
		}
		log.Printf("No return date found for title %s: %v", title.title, err)
	}
	return "", errors.New("No return date found")
}
