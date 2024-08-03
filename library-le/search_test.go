package libraryle

import (
	"testing"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	. "github.com/stretchr/testify/assert"
)

func TestParseGameCopiesResult(t *testing.T) {
	testResponse := loadTestData("testdata/game_copies_example.html")
	games := parseMediaCopiesPage("Monster Hunter Rise", testResponse)
	Equal(t, 4, len(games))

	mediaEqualTo(t, games[0], "Monster Hunter Rise", "Stadtbibliothek", false)
	mediaEqualTo(t, games[1], "Monster Hunter Rise", "Stadtbibliothek", false)
	mediaEqualTo(t, games[2], "Monster Hunter Rise", "Bibliothek Südvorstadt", true)
	mediaEqualTo(t, games[3], "Monster Hunter Rise", "Bibliothek Gohlis", false)
}

func mediaEqualTo(t *testing.T, media domain.Media, exptTitle string, exptBranch string, exptAvalia bool) {
	Equal(t, exptTitle, media.Title)
	Equal(t, exptBranch, media.Branch)
	Equal(t, exptAvalia, media.IsAvailable)
}

func TestParseMovieCopiesResult(t *testing.T) {
	testResponse := loadTestData("testdata/movie_copies_example.html")
	movies := parseMediaCopiesPage("Terminator - Genesis", testResponse)
	Equal(t, 6, len(movies))

	available := 0
	for _, movie := range movies {
		if movie.IsAvailable {
			available++
		}
	}
	Equal(t, 2, available)
}

func TestParseSearchResultMovies(t *testing.T) {
	testResponse := loadTestData("testdata/movie_search_result.html")
	results := extractTitles(testResponse)
	Equal(t, 3, len(results))

	Equal(t, "Der Clou", results[0].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=1&identifier=-1_FT_613132921", results[0].resultUrl)

	Equal(t, "Der Clou", results[1].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=2&identifier=-1_FT_613132921", results[1].resultUrl)

	Equal(t, "Der Clou", results[2].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=3&identifier=-1_FT_613132921", results[2].resultUrl)

}

func TestParseSearchResultGames(t *testing.T) {
	testResponse := loadTestData("testdata/game_search_result.html")
	results := extractTitles(testResponse)
	Equal(t, 3, len(results))

	Equal(t, "Monster hunter generations ultimate", results[0].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=1&identifier=-1_FT_256756711", results[0].resultUrl)

	Equal(t, "Monster hunter rise", results[1].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=2&identifier=-1_FT_256756711", results[1].resultUrl)

	Equal(t, "Monster Hunter - Stories 2. Wings of Ruin", results[2].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=3&identifier=-1_FT_256756711", results[2].resultUrl)
}

func TestClearTitle(t *testing.T) {
	Equal(t, "Terminator", clearTitle("Terminator [Bildtonträger]"))
	Equal(t, "Mad Max - Fury Road", clearTitle("Mad Max - Fury Road [blu-ray]"))
}

func TestRemoveBranchSuffix(t *testing.T) {
	Equal(t, "Bibliothek Gohlis", removeBranchSuffix("Bibliothek Gohlis / Erwachsenenbibliothek"))
	Equal(t, "Bibliothek Grünau-Nord", removeBranchSuffix("Bibliothek Grünau-Nord / Erwachsenenbibliothek"))
	Equal(t, "Fahrbibliothek", removeBranchSuffix("Fahrbibliothek"))
	Equal(t, "", removeBranchSuffix(""))
}
