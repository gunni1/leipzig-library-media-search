package libraryle

import (
	"bufio"
	"log"
	"os"
	"testing"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	. "github.com/stretchr/testify/assert"
)

func TestAvailability(t *testing.T) {
	file, fileErr := os.Open("searchresult_example.html")
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	fileReader := bufio.NewReader(file)
	games, _ := parseSearchResult(fileReader)

	True(t, hasElement(games, "Spiel2"))
	False(t, hasElement(games, "Spiel1"))

}

func hasElement(games []domain.Game, title string) bool {
	for _, game := range games {
		if game.Title == title {
			return true
		}
	}
	return false
}
