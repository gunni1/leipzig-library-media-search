package libraryle

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
	games := parseMediaCopiesPage("Monster Hunter Rise", asDoc(testResponse))
	Equal(t, 4, len(games))

	mediaEqualTo(t, games[0], "Monster Hunter Rise", "Stadtbibliothek", false)
	mediaEqualTo(t, games[1], "Monster Hunter Rise", "Stadtbibliothek", false)
	mediaEqualTo(t, games[2], "Monster Hunter Rise", "Bibliothek Südvorstadt", true)
	mediaEqualTo(t, games[3], "Monster Hunter Rise", "Bibliothek Gohlis", false)
}

// Parser test: parseMediaCopiesPage with movies
func TestParseMovieCopiesResult(t *testing.T) {
	testResponse := loadTestData("testdata/movie_copies_example.html")
	movies := parseMediaCopiesPage("Terminator - Genesis", asDoc(testResponse))
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
