package libraryle

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-media-search/catalog"
	"github.com/gunni1/leipzig-library-media-search/domain"
)

// Verify that Client implements catalog.Client interface
var _ catalog.Client = (*Client)(nil)

// FindMovies searches for movies by title across all library branches
func (libClient Client) FindMovies(title string) ([]domain.Media, error) {
	if err := libClient.newSession(); err != nil {
		return nil, fmt.Errorf("FindMovies: failed to create session: %w", err)
	}

	searchRequest := NewMovieSearchRequest(title, 0, libClient.session)
	searchResponse, err := httpClient.Do(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("FindMovies: HTTP request failed: %w", err)
	}
	defer searchResponse.Body.Close()

	doc, docErr := goquery.NewDocumentFromReader(searchResponse.Body)
	if docErr != nil {
		return nil, fmt.Errorf("FindMovies: could not parse response: %w", docErr)
	}

	if isSingleResultPage(doc) {
		return parseMediaCopiesPage(title, doc), nil
	}

	resultTitles := extractTitles(doc)
	movies := make([]domain.Media, 0)

	for _, resultTitle := range resultTitles {
		copies, err := loadMediaCopies(resultTitle, libClient.session)
		if err != nil {
			log.Printf("FindMovies: failed to load copies for %s: %v", resultTitle.title, err)
			continue
		}
		movies = append(movies, copies...)
	}

	return movies, nil
}

// FindGames searches for games by title and platform across all library branches
func (libClient Client) FindGames(title string, platform string) ([]domain.Media, error) {
	if err := libClient.newSession(); err != nil {
		return nil, fmt.Errorf("FindGames: failed to create session: %w", err)
	}

	searchRequest := NewGameSearchRequest(title, platform, 0, libClient.session)
	searchResponse, err := httpClient.Do(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("FindGames: HTTP request failed: %w", err)
	}
	defer searchResponse.Body.Close()

	doc, docErr := goquery.NewDocumentFromReader(searchResponse.Body)
	if docErr != nil {
		return nil, fmt.Errorf("FindGames: could not parse response: %w", docErr)
	}

	if isSingleResultPage(doc) {
		return parseMediaCopiesPage(title, doc), nil
	}

	resultTitles := extractTitles(doc)
	games := make([]domain.Media, 0)

	for _, resultTitle := range resultTitles {
		copies, err := loadMediaCopies(resultTitle, libClient.session)
		if err != nil {
			log.Printf("FindGames: failed to load copies for %s: %v", resultTitle.title, err)
			continue
		}
		games = append(games, copies...)
	}

	return games, nil
}

// FindAvailableGames lists all games available in a specific branch for a platform
func (libClient Client) FindAvailableGames(branchCode int, platform string) ([]domain.Game, error) {
	if err := libClient.newSession(); err != nil {
		return nil, fmt.Errorf("FindAvailableGames: failed to create session: %w", err)
	}

	request := NewGameIndexRequest(branchCode, platform, libClient.session)
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("FindAvailableGames: HTTP request failed: %w", err)
	}
	defer response.Body.Close()

	games, parseErr := parseGameSearchResult(response.Body)
	if parseErr != nil {
		return nil, fmt.Errorf("FindAvailableGames: could not parse response: %w", parseErr)
	}

	return games, nil
}

// RetrieveReturnDate returns the next available return date for a title in a specific branch
func (libClient Client) RetrieveReturnDate(branchCode int, platform string, title string) (string, error) {
	if err := libClient.newSession(); err != nil {
		return "", fmt.Errorf("RetrieveReturnDate: failed to create session: %w", err)
	}

	request := NewReturnDateRequest(title, platform, branchCode, libClient.session)
	searchResponse, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("RetrieveReturnDate: HTTP request failed: %w", err)
	}
	defer searchResponse.Body.Close()

	doc, docErr := goquery.NewDocumentFromReader(searchResponse.Body)
	if docErr != nil {
		return "", fmt.Errorf("RetrieveReturnDate: could not parse response: %w", docErr)
	}

	if isSingleResultPage(doc) {
		return findReturnDateInCopiesPage(doc)
	}

	resultTitles := extractTitles(doc)
	exactMatchTitles := filterExactTitle(title, resultTitles)
	return loadMediaReturnDate(exactMatchTitles, libClient.session)
}

// loadMediaCopies fetches and parses the copies page for a single search result
func loadMediaCopies(result searchResult, libSession webOpacSession) ([]domain.Media, error) {
	request := createRequest(libSession, result.resultUrl)

	mediaResponse, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("loadMediaCopies: HTTP request failed: %w", err)
	}
	defer mediaResponse.Body.Close()

	doc, docErr := goquery.NewDocumentFromReader(mediaResponse.Body)
	if docErr != nil {
		return nil, fmt.Errorf("loadMediaCopies: could not parse response: %w", docErr)
	}

	return parseMediaCopiesPage(result.title, doc), nil
}

// loadMediaReturnDate fetches and parses return dates for multiple search results
// Returns the first successful date found
func loadMediaReturnDate(titles []searchResult, libSession webOpacSession) (string, error) {
	for _, title := range titles {
		returnDate, err := loadReturnDate(title, libSession)
		if err == nil {
			return returnDate, nil
		}
		log.Printf("loadMediaReturnDate: could not get return date for %s: %v", title.title, err)
	}
	return "", fmt.Errorf("loadMediaReturnDate: no return date found for any result")
}

// loadReturnDate fetches the return date for a single search result
func loadReturnDate(result searchResult, libSession webOpacSession) (string, error) {
	request := createRequest(libSession, result.resultUrl)

	mediaResponse, err := httpClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("loadReturnDate: HTTP request failed: %w", err)
	}
	defer mediaResponse.Body.Close()

	doc, docErr := goquery.NewDocumentFromReader(mediaResponse.Body)
	if docErr != nil {
		return "", fmt.Errorf("loadReturnDate: could not parse response: %w", docErr)
	}

	return findReturnDateInCopiesPage(doc)
}
