package libraryle

import (
	"bufio"
	"io"
	"log"
	"os"
	"testing"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	. "github.com/stretchr/testify/assert"
)

func TestAvailability(t *testing.T) {
	fileReader := loadTestData("game_search_example.html")

	games, _ := parseGameSearchResult(fileReader)

	True(t, hasElement(games, "Spiel2"))
	False(t, hasElement(games, "Spiel1"))

}

func TestParseMovieCopiesResult(t *testing.T) {
	testResponse := loadTestData("movie_copies_example.html")
	movies := parseMovieCopiesPage(testResponse)
	Equal(t, 6, len(movies))
}

func loadTestData(filePath string) io.Reader {
	file, fileErr := os.Open(filePath)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	return bufio.NewReader(file)
}

func hasElement(games []domain.Game, title string) bool {
	for _, game := range games {
		if game.Title == title {
			return true
		}
	}
	return false
}
