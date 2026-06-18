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
func parseMediaCopiesPage(title string, doc *goquery.Document) []domain.Media {
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

	return media
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
