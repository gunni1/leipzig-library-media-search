package libraryle

import (
	"fmt"
	"log"

	"github.com/gunni1/leipzig-library-media-search/domain"
)

func (libClient Client) FindAvailabelGames(branchCode int, platform string) []domain.Game {
	sessionErr := libClient.newSession()
	if sessionErr != nil {
		fmt.Println(sessionErr)
		return nil
	}
	request := NewGameIndexRequest(branchCode, platform, libClient.session)
	response, err := httpClient.Do(request)
	if err != nil {
		log.Println("error during search")
		return nil
	}
	defer response.Body.Close()

	games, parseResultErr := parseGameSearchResult(response.Body)
	if parseResultErr != nil {
		log.Printf("parse error during game index search: %v", parseResultErr)
		return nil
	}
	return games
}
