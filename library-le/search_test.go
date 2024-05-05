package libraryle

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestParseMovieCopiesResult(t *testing.T) {
	testResponse := loadTestData("testdata/movie_copies_example.html")
	movies := parseMovieCopiesPage("Terminator - Genesis", testResponse)
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
	results := parseMediaSearch(testResponse)
	Equal(t, 3, len(results))

	Equal(t, "Der Clou", results[0].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=1&identifier=-1_FT_613132921", results[0].resultUrl)

	Equal(t, "Der Clou", results[1].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=2&identifier=-1_FT_613132921", results[1].resultUrl)

	Equal(t, "Der Clou [Blu-ray]", results[2].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=3&identifier=-1_FT_613132921", results[2].resultUrl)

}

func TestParseSearchResultGames(t *testing.T) {
	testResponse := loadTestData("testdata/game_search_result.html")
	results := parseMediaSearch(testResponse)
	Equal(t, 3, len(results))

	Equal(t, "Monster hunter generations ultimate", results[0].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=1&identifier=-1_FT_256756711", results[0].resultUrl)

	Equal(t, "Monster hunter rise", results[1].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=2&identifier=-1_FT_256756711", results[1].resultUrl)

	Equal(t, "Monster Hunter - Stories 2. Wings of Ruin", results[2].title)
	Equal(t, "/webOPACClient/singleHit.do?methodToCall=showHit&curPos=3&identifier=-1_FT_256756711", results[2].resultUrl)

}
