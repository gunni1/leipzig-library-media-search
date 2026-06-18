# Catalog Abstraction Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a Go interface that hides all HTML processing behind a catalog API, enabling testability, clean separation of concerns, and future extensibility.

**Architecture:** New `catalog/` package defines a `Client` interface. `library-le/` is restructured internally into `parser.go` (pure HTML logic), `opac.go` (implements the interface and orchestrates), and supporting files. `web/` and `cmd/cli` depend on the abstraction.

**Tech Stack:** Go, goquery (HTML parsing), testify (assertions), httptest (HTTP mocking).

---

## File Structure

**Created:**
- `catalog/client.go` — interface definition

**Modified:**
- `library-le/parser.go` — new file, all HTML parsing logic extracted here
- `library-le/opac.go` — new file, implements `catalog.Client`, wires session + HTTP + parser
- `library-le/client.go` — session management stays, no logic changes
- `library-le/request.go` — unchanged, already clean
- `library-le/branches.go` — unchanged
- `library-le/search.go` — deleted after logic is extracted
- `library-le/game-index.go` — deleted after logic is extracted
- `library-le/search_test.go` → migrated to `library-le/parser_test.go`
- `library-le/game-index_test.go` → migrated to `library-le/parser_test.go`
- `library-le/client_test.go` — updated to work with new structure
- `web/server.go` — update to use `catalog.Client` interface instead of `libClient.Client` directly
- `cmd/cli/main.go` — update to use `catalog.Client` interface

---

## Task 1: Create `catalog/client.go` interface

**Files:**
- Create: `catalog/client.go`

- [ ] **Step 1: Create the catalog package directory**

```bash
mkdir -p /Users/guntram/git/leipzig-library-media-search/catalog
```

- [ ] **Step 2: Write the catalog.Client interface**

Create `/Users/guntram/git/leipzig-library-media-search/catalog/client.go`:

```go
package catalog

import (
	"github.com/gunni1/leipzig-library-media-search/domain"
)

// Client defines the interface for querying the library catalog.
// Implementations hide the details of HTTP session management,
// request building, and HTML parsing.
type Client interface {
	// FindMovies searches for movies by title across all library branches.
	// Returns a slice of Media or an error if the search fails.
	// Empty results (nil error, empty slice) are different from errors.
	FindMovies(title string) ([]domain.Media, error)

	// FindGames searches for games by title and platform across all library branches.
	// Returns a slice of Media or an error if the search fails.
	FindGames(title, platform string) ([]domain.Media, error)

	// FindAvailableGames lists all games available in a specific branch for a platform.
	// Returns a slice of Game or an error if the search fails.
	FindAvailableGames(branchCode int, platform string) ([]domain.Game, error)

	// RetrieveReturnDate returns the next available return date for a title in a specific branch.
	// Returns the date as a string (format: DD.MM.YYYY) or an error if not found.
	RetrieveReturnDate(branchCode int, platform, title string) (string, error)
}
```

- [ ] **Step 3: Commit**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git add catalog/client.go
git commit -m "feat: add catalog.Client interface"
```

---

## Task 2: Extract HTML parsing logic into `library-le/parser.go`

**Files:**
- Create: `library-le/parser.go`
- Source: Extract functions from `library-le/search.go` and `library-le/game-index.go`

- [ ] **Step 1: Create parser.go and copy all parsing functions**

Create `/Users/guntram/git/leipzig-library-media-search/library-le/parser.go` with:

```go
package libraryle

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-media-search/domain"
	"github.com/pkg/errors"
)

const (
	copiesSelector   string = "#tab-content > div > div:nth-child(n+2)"
	packageSelector  string = "div.results-teaser > div > div > ul > li:nth-child(4)" //Umfang
	keyWordsSelector string = "div.results-teaser > div > div > ul > li:nth-child(5)" //Schlagwort

	resultItemSelector   string = "h2[class^=recordtitle]"
	titleSelector        string = "a[href^='/webOPACClient/singleHit']"
	availabilitySelector string = "span[class^=textgruen]"
)

var (
	dateRx          = regexp.MustCompile(`\d{2}\.\d{2}\.\d{4}`)
	platformsRx     = regexp.MustCompile(`playstation|x-box|switch`)
	titleBracketsRx = regexp.MustCompile(`\[.*?\]`)
)

// Private struct to hold search result data (title + URL)
type searchResult struct {
	title     string
	resultUrl string
}

// isSingleResultPage checks if the response is a single result page vs a search results list
func isSingleResultPage(doc *goquery.Document) bool {
	pageTitle := doc.Find("title").Text()
	return strings.TrimSpace(pageTitle) == "Einzeltreffer"
}

// extractTitles parses a search results overview page and returns all found titles with their URLs
func extractTitles(doc *goquery.Document) []searchResult {
	titles := make([]searchResult, 0)
	doc.Find(resultItemSelector).Each(func(i int, resultItem *goquery.Selection) {
		title := clearTitle(resultItem.Find(titleSelector).Text())
		resultUrl, _ := resultItem.Find(titleSelector).Attr("href")
		titles = append(titles, searchResult{title: title, resultUrl: resultUrl})
	})
	return titles
}

// parseMediaCopiesPage parses the details page for a single media item,
// returning copies available in each library branch
func parseMediaCopiesPage(title string, doc *goquery.Document) ([]domain.Media, error) {
	media := make([]domain.Media, 0)

	platform := determinePlatform(doc)
	if platform == "" {
		log.Printf("Could not determine platform for title: %s\n", title)
	}

	doc.Find(copiesSelector).Each(func(i int, copy *goquery.Selection) {
		branch := copy.Find("div.col-12.col-md-4.my-md-2 > b").Text()
		title := strings.TrimSpace(doc.Find("#middle > div.box > div > h2").Text())
		status := isMediaAvailable(copy)
		media = append(media, domain.Media{
			Title:       title,
			Branch:      removeBranchSuffix(branch),
			Platform:    platform,
			IsAvailable: status,
		})
	})

	return media, nil
}

// parseGameSearchResult parses a game index search result and returns only available games
func parseGameSearchResult(searchResult io.Reader) ([]domain.Game, error) {
	doc, docErr := goquery.NewDocumentFromReader(searchResult)
	if docErr != nil {
		return nil, fmt.Errorf("could not parse game search result: %w", docErr)
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

// findReturnDateInCopiesPage searches the copies page for the first available return date
func findReturnDateInCopiesPage(doc *goquery.Document) (string, error) {
	returnDate := ""
	doc.Find(copiesSelector).Each(func(i int, copy *goquery.Selection) {
		rentalStateLink := copy.Find("div:nth-child(5) > div > a")
		dateStr, findErr := extractDate(rentalStateLink.Text())
		if findErr == nil {
			returnDate = dateStr
		}
	})

	if returnDate != "" {
		return returnDate, nil
	}
	return "", errors.New("found no copy with a return date")
}

// determinePlatform extracts the platform (DVD, Blu-ray, PlayStation, Xbox, Switch) from the page
func determinePlatform(doc *goquery.Document) string {
	keyWords := strings.ToLower(doc.Find(keyWordsSelector).Text())
	if platformsRx.MatchString(keyWords) {
		if strings.Contains(keyWords, "playstation") {
			return "playstation"
		} else if strings.Contains(keyWords, "x-box") {
			return "xbox"
		} else if strings.Contains(keyWords, "switch") {
			return "switch"
		}
	} else {
		moviePlatform := strings.ToLower(doc.Find(packageSelector).Text())
		if strings.Contains(moviePlatform, "dvd") {
			return "dvd"
		} else if strings.Contains(moviePlatform, "blu-ray") {
			return "bluray"
		}
	}
	return ""
}

// isMediaAvailable checks if a single copy is available for checkout
func isMediaAvailable(copy *goquery.Selection) bool {
	rentalStateLink := copy.Find("div:nth-child(5) > div > a")
	// Link indicates a rented state (can reserve a copy)
	if rentalStateLink.Length() != 0 {
		return false
	}
	statusText := copy.Find("div:nth-child(5)").Text()
	return strings.Contains(statusText, "ausleihbar") || strings.Contains(statusText, "frei")
}

// isGameAvailable checks if a game search result indicates availability
func isGameAvailable(searchHitNode *goquery.Selection) bool {
	return searchHitNode.Find(availabilitySelector).Length() > 0
}

// extractDate extracts a date string in DD.MM.YYYY format from text
func extractDate(text string) (string, error) {
	date := dateRx.FindString(text)
	if date == "" {
		return "", fmt.Errorf("no date found in: %s", text)
	}
	return date, nil
}

// clearTitle removes additional media information in square brackets from a title
func clearTitle(title string) string {
	return strings.TrimSpace(titleBracketsRx.ReplaceAllString(title, ""))
}

// removeBranchSuffix removes the location detail suffix from branch names
func removeBranchSuffix(branchName string) string {
	return strings.TrimSpace(strings.Split(branchName, "/")[0])
}

// filterExactTitle filters search results to exact title matches
func filterExactTitle(title string, results []searchResult) []searchResult {
	filtered := make([]searchResult, 0)
	for _, result := range results {
		if result.title == title {
			filtered = append(filtered, result)
		}
	}
	return filtered
}
```

- [ ] **Step 2: Commit**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git add library-le/parser.go
git commit -m "feat: extract HTML parsing logic into parser.go"
```

---

## Task 3: Create `library-le/opac.go` implementing `catalog.Client`

**Files:**
- Create: `library-le/opac.go`
- This orchestrates session management, HTTP requests, and calls parser functions

- [ ] **Step 1: Write opac.go**

Create `/Users/guntram/git/leipzig-library-media-search/library-le/opac.go`:

```go
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
		return parseMediaCopiesPage(title, doc)
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
		return parseMediaCopiesPage(title, doc)
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

	return parseMediaCopiesPage(result.title, doc)
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
```

- [ ] **Step 2: Commit**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git add library-le/opac.go
git commit -m "feat: implement catalog.Client in opac.go"
```

---

## Task 4: Migrate existing parser tests to `library-le/parser_test.go`

**Files:**
- Create: `library-le/parser_test.go` (combines search_test.go and game-index_test.go parser tests)
- Keep: `library-le/client_test.go` (request/session tests)

- [ ] **Step 1: Create parser_test.go with all parser tests**

Create `/Users/guntram/git/leipzig-library-media-search/library-le/parser_test.go`:

```go
package libraryle

import (
	"io"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-media-search/domain"
	. "github.com/stretchr/testify/assert"
)

// Helper: load test data file
func loadTestData(filePath string) io.Reader {
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		panic(fileErr)
	}
	return bufio.NewReader(file)
}

// Helper: convert reader to goquery Document
func asDoc(reader io.Reader) *goquery.Document {
	doc, _ := goquery.NewDocumentFromReader(reader)
	return doc
}

// Parser test: parseMediaCopiesPage with games
func TestParseGameCopiesResult(t *testing.T) {
	testResponse := loadTestData("testdata/game_copies_example.html")
	games, err := parseMediaCopiesPage("Monster Hunter Rise", asDoc(testResponse))
	Nil(t, err)
	Equal(t, 4, len(games))

	mediaEqualTo(t, games[0], "Monster Hunter Rise", "Stadtbibliothek", false)
	mediaEqualTo(t, games[1], "Monster Hunter Rise", "Stadtbibliothek", false)
	mediaEqualTo(t, games[2], "Monster Hunter Rise", "Bibliothek Südvorstadt", true)
	mediaEqualTo(t, games[3], "Monster Hunter Rise", "Bibliothek Gohlis", false)
}

// Parser test: parseMediaCopiesPage with movies
func TestParseMovieCopiesResult(t *testing.T) {
	testResponse := loadTestData("testdata/movie_copies_example.html")
	movies, err := parseMediaCopiesPage("Terminator - Genesis", asDoc(testResponse))
	Nil(t, err)
	Equal(t, 6, len(movies))

	available := 0
	for _, movie := range movies {
		if movie.IsAvailable {
			available++
		}
	}
	Equal(t, 2, available)
}

// Parser test: extractTitles from movie search results
func TestParseSearchResultMovies(t *testing.T) {
	testResponse := loadTestData("testdata/movie_search_result.html")
	results := extractTitles(asDoc(testResponse))
	Equal(t, 3, len(results))

	Equal(t, "Der Clou", results[0].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=1&identifier=-1_FT_613132921", results[0].resultUrl)

	Equal(t, "Der Clou", results[1].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=2&identifier=-1_FT_613132921", results[1].resultUrl)

	Equal(t, "Der Clou", results[2].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=3&identifier=-1_FT_613132921", results[2].resultUrl)
}

// Parser test: extractTitles from game search results
func TestParseSearchResultGames(t *testing.T) {
	testResponse := loadTestData("testdata/game_search_result.html")
	results := extractTitles(asDoc(testResponse))
	Equal(t, 3, len(results))

	Equal(t, "Monster hunter generations ultimate", results[0].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=1&identifier=-1_FT_256756711", results[0].resultUrl)

	Equal(t, "Monster hunter rise", results[1].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=2&identifier=-1_FT_256756711", results[1].resultUrl)

	Equal(t, "Monster Hunter - Stories 2. Wings of Ruin", results[2].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=3&identifier=-1_FT_256756711", results[2].resultUrl)
}

// Parser test: parseGameSearchResult
func TestParseGameSearchResult(t *testing.T) {
	fileReader := loadTestData("testdata/game_search_example.html")
	games, err := parseGameSearchResult(fileReader)
	Nil(t, err)

	True(t, hasElement(games, "Spiel2"))
	False(t, hasElement(games, "Spiel1"))
}

// Parser test: clearTitle
func TestClearTitle(t *testing.T) {
	Equal(t, "Terminator", clearTitle("Terminator [Bildtonträger]"))
	Equal(t, "Mad Max - Fury Road", clearTitle("Mad Max - Fury Road [blu-ray]"))
}

// Parser test: removeBranchSuffix
func TestRemoveBranchSuffix(t *testing.T) {
	Equal(t, "Bibliothek Gohlis", removeBranchSuffix("Bibliothek Gohlis / Erwachsenenbibliothek"))
	Equal(t, "Bibliothek Grünau-Nord", removeBranchSuffix("Bibliothek Grünau-Nord / Erwachsenenbibliothek"))
	Equal(t, "Fahrbibliothek", removeBranchSuffix("Fahrbibliothek"))
	Equal(t, "", removeBranchSuffix(""))
}

// Parser test: determinePlatform
func TestDetermPlatform(t *testing.T) {
	Equal(t, "xbox", determinePlatform(asDoc(loadTestData("testdata/determ_platform_xbox.html"))))
	Equal(t, "playstation", determinePlatform(asDoc(loadTestData("testdata/determ_platform_ps.html"))))
	Equal(t, "switch", determinePlatform(asDoc(loadTestData("testdata/determ_platform_switch.html"))))
	Equal(t, "dvd", determinePlatform(asDoc(loadTestData("testdata/determ_platform_dvd.html"))))
	Equal(t, "bluray", determinePlatform(asDoc(loadTestData("testdata/determ_platform_bluray.html"))))
}

// Parser test: filterExactTitle
func TestFilterSearchResult(t *testing.T) {
	search := []searchResult{
		{title: "Terminator"},
		{title: "Terminator 2"},
	}
	filtered := filterExactTitle("Terminator", search)
	Equal(t, 1, len(filtered))
	Equal(t, "Terminator", filtered[0].title)
}

// Parser test: extractDate
func TestExtractDate(t *testing.T) {
	date, err := extractDate("Today is the 20.08.2024.")
	Equal(t, "20.08.2024", date)
	Nil(t, err)

	_, err = extractDate("Whops, this date has a formatting issue: 11.11,2011")
	NotNil(t, err)
}

// Parser test: isSingleResultPage TRUE
func TestIsSinglePageResultTRUE(t *testing.T) {
	data := strings.NewReader("<html><head><title>   \n Einzeltreffer   \n </title></head></html>")
	result := isSingleResultPage(asDoc(data))
	True(t, result)
}

// Parser test: isSingleResultPage FALSE
func TestIsSinglePageResultFALSE(t *testing.T) {
	data := strings.NewReader("<html><head><title> Trefferliste </title></head></html>")
	result := isSingleResultPage(asDoc(data))
	False(t, result)
}

// Helper: check if game exists in list
func hasElement(games []domain.Game, title string) bool {
	for _, game := range games {
		if game.Title == title {
			return true
		}
	}
	return false
}

// Helper: assert media fields match
func mediaEqualTo(t *testing.T, media domain.Media, exptTitle string, exptBranch string, exptAvail bool) {
	Equal(t, exptTitle, media.Title)
	Equal(t, exptBranch, media.Branch)
	Equal(t, exptAvail, media.IsAvailable)
}
```

**Note:** Add imports at the top:
```go
import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gunni1/leipzig-library-media-search/domain"
	. "github.com/stretchr/testify/assert"
)
```

- [ ] **Step 2: Run parser tests to verify they pass**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go test ./library-le -run TestParse -v
```

Expected output: All parser tests pass (TestParseGameCopiesResult, TestParseMovieCopiesResult, etc.)

- [ ] **Step 3: Commit**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git add library-le/parser_test.go
git commit -m "test: migrate parser tests to parser_test.go"
```

---

## Task 5: Update `library-le/client_test.go` to verify interface compliance

**Files:**
- Modify: `library-le/client_test.go`

- [ ] **Step 1: Add interface compliance check at the top of the test file**

In `/Users/guntram/git/leipzig-library-media-search/library-le/client_test.go`, add this import and compliance check:

```go
import (
	// ... existing imports ...
	"github.com/gunni1/leipzig-library-media-search/catalog"
)

// Verify that Client implements catalog.Client
var _ catalog.Client = (*Client)(nil)
```

(Add this right after the package declaration and before existing code)

- [ ] **Step 2: Run client tests to verify they still pass**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go test ./library-le/client_test.go -v
```

Expected output: All existing client tests pass

- [ ] **Step 3: Commit**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git add library-le/client_test.go
git commit -m "test: add catalog.Client interface compliance check"
```

---

## Task 6: Update `web/server.go` to use `catalog.Client` interface

**Files:**
- Modify: `web/server.go`

- [ ] **Step 1: Add catalog import**

In `/Users/guntram/git/leipzig-library-media-search/web/server.go`, add to imports:

```go
import (
	// ... existing imports ...
	"github.com/gunni1/leipzig-library-media-search/catalog"
)
```

- [ ] **Step 2: Update gameSearchHandler to accept interface**

Find this function in `server.go` (around line 86):

```go
func gameSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "game search")
	title := strings.ToLower(request.PostFormValue("title"))
	platform := strings.ToLower(request.PostFormValue("platform"))
	showNotAvailable := strings.ToLower(request.PostFormValue("showNotAvailable")) == "true"

	client := libClient.Client{}
	games := client.FindGames(title, platform)
	if !showNotAvailable {
		games = filterAvailable(games)
	}
	renderMediaResults(games, domain.GAME, respWriter, request)
}
```

Replace it with:

```go
func gameSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "game search")
	title := strings.ToLower(request.PostFormValue("title"))
	platform := strings.ToLower(request.PostFormValue("platform"))
	showNotAvailable := strings.ToLower(request.PostFormValue("showNotAvailable")) == "true"

	client := libClient.NewClientWithSession()
	games, err := client.FindGames(title, platform)
	if err != nil {
		log.Printf("gameSearchHandler: %v", err)
		fmt.Fprint(respWriter, "<p>Suche fehlgeschlagen.</p>")
		return
	}
	if !showNotAvailable {
		games = filterAvailable(games)
	}
	renderMediaResults(games, domain.GAME, respWriter, request)
}
```

- [ ] **Step 3: Update movieSearchHandler to accept interface**

Find this function (around line 100):

```go
func movieSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "movie search")
	title := strings.ToLower(request.PostFormValue("movie-title"))
	showNotAvailable := strings.ToLower(request.PostFormValue("showNotAvailable")) == "true"

	client := libClient.Client{}
	movies := client.FindMovies(title)
	if !showNotAvailable {
		movies = filterAvailable(movies)
	}
	renderMediaResults(movies, domain.MOVIE, respWriter, request)
}
```

Replace it with:

```go
func movieSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "movie search")
	title := strings.ToLower(request.PostFormValue("movie-title"))
	showNotAvailable := strings.ToLower(request.PostFormValue("showNotAvailable")) == "true"

	client := libClient.NewClientWithSession()
	movies, err := client.FindMovies(title)
	if err != nil {
		log.Printf("movieSearchHandler: %v", err)
		fmt.Fprint(respWriter, "<p>Suche fehlgeschlagen.</p>")
		return
	}
	if !showNotAvailable {
		movies = filterAvailable(movies)
	}
	renderMediaResults(movies, domain.MOVIE, respWriter, request)
}
```

- [ ] **Step 4: Update gameIndexHandler to use error handling**

Find this function (around line 326):

```go
func gameIndexHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "game index")
	branch := strings.ToLower(request.PostFormValue("branch"))
	platform := strings.ToLower(request.PostFormValue("platform"))
	branchCode, exists := libClient.GetBranchCode(branch)
	if !exists {
		log.Printf("Requested branch: %s does not exists.", branch)
		return
	}
	client := libClient.Client{}
	games := client.FindAvailabelGames(branchCode, platform)

	if len(games) == 0 {
		fmt.Fprint(respWriter, "<p>Es wurden keine ausleihbaren Titel gefunden.</p>")
		return
	}
	data := map[string][]domain.Game{
		"Items": games,
	}
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/item-list.html"))
	templ.Execute(respWriter, data)
}
```

Replace it with:

```go
func gameIndexHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "game index")
	branch := strings.ToLower(request.PostFormValue("branch"))
	platform := strings.ToLower(request.PostFormValue("platform"))
	branchCode, exists := libClient.GetBranchCode(branch)
	if !exists {
		log.Printf("Requested branch: %s does not exist.", branch)
		fmt.Fprint(respWriter, "<p>Bibliothek nicht gefunden.</p>")
		return
	}
	client := libClient.NewClientWithSession()
	games, err := client.FindAvailableGames(branchCode, platform)
	if err != nil {
		log.Printf("gameIndexHandler: %v", err)
		fmt.Fprint(respWriter, "<p>Suche fehlgeschlagen.</p>")
		return
	}

	if len(games) == 0 {
		fmt.Fprint(respWriter, "<p>Es wurden keine ausleihbaren Titel gefunden.</p>")
		return
	}
	data := map[string][]domain.Game{
		"Items": games,
	}
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/item-list.html"))
	templ.Execute(respWriter, data)
}
```

- [ ] **Step 5: Update watchlistCheckHandler to handle errors**

Find this function (around line 131):

```go
func watchlistCheckHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "watchlist check")
	title := request.URL.Query().Get("title")
	platform := request.URL.Query().Get("platform")
	mediaType := request.URL.Query().Get("type")

	client := libClient.Client{}
	var medias []domain.Media
	if mediaType == domain.MOVIE {
		medias = client.FindMovies(title)
	} else {
		medias = client.FindGames(title, platform)
	}

	// Filter to exact title match (library search is fuzzy)
	titleLower := strings.ToLower(title)
	filtered := make([]domain.Media, 0)
	for _, m := range medias {
		if strings.ToLower(m.Title) == titleLower {
			filtered = append(filtered, m)
		}
	}

	byBranch := arrangeByBranch(filtered)
	data := MediaTemplateData{
		Branches:  byBranch,
		MediaType: mediaType,
	}
	templ, err := template.New("watchlist-check.html").Funcs(template.FuncMap{
		"encodeBranch": encodeBranch,
	}).ParseFS(htmlTemplates, "templates/watchlist-check.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err := templ.Execute(respWriter, data); err != nil {
		log.Println(err)
	}
}
```

Replace it with:

```go
func watchlistCheckHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "watchlist check")
	title := request.URL.Query().Get("title")
	platform := request.URL.Query().Get("platform")
	mediaType := request.URL.Query().Get("type")

	client := libClient.NewClientWithSession()
	var medias []domain.Media
	var err error
	if mediaType == domain.MOVIE {
		medias, err = client.FindMovies(title)
	} else {
		medias, err = client.FindGames(title, platform)
	}

	if err != nil {
		log.Printf("watchlistCheckHandler: %v", err)
		http.Error(respWriter, "search failed", http.StatusInternalServerError)
		return
	}

	// Filter to exact title match (library search is fuzzy)
	titleLower := strings.ToLower(title)
	filtered := make([]domain.Media, 0)
	for _, m := range medias {
		if strings.ToLower(m.Title) == titleLower {
			filtered = append(filtered, m)
		}
	}

	byBranch := arrangeByBranch(filtered)
	data := MediaTemplateData{
		Branches:  byBranch,
		MediaType: mediaType,
	}
	templ, err := template.New("watchlist-check.html").Funcs(template.FuncMap{
		"encodeBranch": encodeBranch,
	}).ParseFS(htmlTemplates, "templates/watchlist-check.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err := templ.Execute(respWriter, data); err != nil {
		log.Println(err)
	}
}
```

- [ ] **Step 6: Run web tests to verify they still pass**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go test ./web -v
```

Expected output: web tests pass

- [ ] **Step 7: Commit**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git add web/server.go
git commit -m "refactor: update web handlers to use catalog.Client interface and handle errors"
```

---

## Task 7: Update `cmd/cli/main.go` to use interface

**Files:**
- Modify: `cmd/cli/main.go`

- [ ] **Step 1: Update imports**

In `/Users/guntram/git/leipzig-library-media-search/cmd/cli/main.go`:

```go
import (
	"flag"
	"fmt"
	"log"

	"github.com/gunni1/leipzig-library-media-search/catalog"
	libClient "github.com/gunni1/leipzig-library-media-search/library-le"
)
```

- [ ] **Step 2: Update main function to use interface and handle errors**

Replace the entire main function with:

```go
func main() {
	searchGame := flag.Bool("game", false, "search for a game")
	searchMovie := flag.Bool("movie", false, "search for a movie")

	titlePtr := flag.String("title", "Terminator", "title to search for")
	platformPtr := flag.String("platform", "Nintendo Switch", "Console platform to list games")

	flag.Parse()

	if *searchGame && *searchMovie || !*searchGame && !*searchMovie {
		log.Fatal("please select either -movie OR -game search flag")
	}

	client := libClient.NewClientWithSession()
	var media []domain.Media
	var err error

	if *searchGame {
		media, err = client.FindGames(*titlePtr, *platformPtr)
	} else {
		media, err = client.FindMovies(*titlePtr)
	}

	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	if len(media) == 0 {
		fmt.Println("No results found")
		return
	}

	for _, result := range media {
		fmt.Printf("%#v\n", result)
	}
}
```

But fix the import — it should import `domain` not `Media`. Correct version:

```go
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gunni1/leipzig-library-media-search/domain"
	libClient "github.com/gunni1/leipzig-library-media-search/library-le"
)

func main() {
	searchGame := flag.Bool("game", false, "search for a game")
	searchMovie := flag.Bool("movie", false, "search for a movie")

	titlePtr := flag.String("title", "Terminator", "title to search for")
	platformPtr := flag.String("platform", "Nintendo Switch", "Console platform to list games")

	flag.Parse()

	if *searchGame && *searchMovie || !*searchGame && !*searchMovie {
		log.Fatal("please select either -movie OR -game search flag")
	}

	client := libClient.NewClientWithSession()
	var media []domain.Media
	var err error

	if *searchGame {
		media, err = client.FindGames(*titlePtr, *platformPtr)
	} else {
		media, err = client.FindMovies(*titlePtr)
	}

	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	if len(media) == 0 {
		fmt.Println("No results found")
		return
	}

	for _, result := range media {
		fmt.Printf("%#v\n", result)
	}
}
```

- [ ] **Step 3: Test the CLI still works**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go build -o /tmp/cli-test ./cmd/cli
echo "Build successful"
```

- [ ] **Step 4: Commit**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git add cmd/cli/main.go
git commit -m "refactor: update CLI to use catalog.Client interface and handle errors"
```

---

## Task 8: Remove old files `search.go` and `game-index.go`

**Files:**
- Delete: `library-le/search.go`
- Delete: `library-le/game-index.go`
- Delete: `library-le/search_test.go` (tests migrated to parser_test.go)
- Delete: `library-le/game-index_test.go` (tests migrated to parser_test.go)

- [ ] **Step 1: Delete old source files**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git rm library-le/search.go library-le/game-index.go library-le/search_test.go library-le/game-index_test.go
```

- [ ] **Step 2: Run all tests to verify nothing broke**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go test ./... -v
```

Expected output: All tests pass (parser_test.go, client_test.go, web tests, etc.)

- [ ] **Step 3: Build the project to verify no import errors**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go build -v ./...
```

Expected output: Build successful with no errors

- [ ] **Step 4: Commit**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git commit -m "refactor: remove old search.go and game-index.go, tests migrated to parser_test.go"
```

---

## Task 9: Verify integration and run full test suite

**Files:**
- No file changes; verification only

- [ ] **Step 1: Run all library-le tests**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go test ./library-le -v
```

Expected output: All parser and client tests pass

- [ ] **Step 2: Run all web tests**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go test ./web -v
```

Expected output: All web tests pass

- [ ] **Step 3: Build entire project**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go build -v
```

Expected output: Main binary builds successfully

- [ ] **Step 4: Verify catalog interface is properly defined**

```bash
cd /Users/guntram/git/leipzig-library-media-search
go doc catalog.Client
```

Expected output: Shows the Client interface with all four methods

- [ ] **Step 5: Check git status is clean**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git status
```

Expected output: `working tree clean`

- [ ] **Step 6: View final commit log**

```bash
cd /Users/guntram/git/leipzig-library-media-search
git log --oneline -10
```

Expected output: Shows all the commits from this plan

- [ ] **Step 7: Done**

All tasks complete. The refactoring is done:
- New `catalog.Client` interface is the public API
- HTML parsing is isolated in `parser.go`
- `opac.go` implements the interface
- `web/` and `cmd/cli` use the interface
- Tests are comprehensive and passing
- Old tangled code is removed
