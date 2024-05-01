package libraryle

import (
	"testing"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	. "github.com/stretchr/testify/assert"
)

func TestAvailability(t *testing.T) {
	fileReader := loadTestData("testdata/game_search_example.html")

	games, _ := parseGameSearchResult(fileReader)

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
