package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/gunni1/leipzig-library-game-stock-api/pkg/domain"
	libClient "github.com/gunni1/leipzig-library-game-stock-api/pkg/library-le"
)

func main() {
	branchPtr := flag.Int("branch", 20, "Branch code of the library")
	platformPtr := flag.String("platform", "Nintendo Switch", "Console platform to list games")
	allBranchesPtr := flag.Bool("all", false, "Search in all Branches")

	flag.Parse()

	client := libClient.Client{}

	var games []domain.Game
	if *allBranchesPtr {
		games = GetAllAvailableGamesPlatform(*platformPtr)
	} else {
		games = client.FindAvailabelGames(*branchPtr, *platformPtr)
	}

	for _, game := range games {
		fmt.Printf("%s (%s)\n", game.Title, game.Branch)
	}

}

func GetAllAvailableGamesPlatform(platform string) []domain.Game {
	branchCodes := libClient.BranchCodeKeys()
	searchResults := make(chan domain.Game)

	wg := &sync.WaitGroup{}
	for _, code := range branchCodes {
		wg.Add(1)
		go getAvailableGames(code, platform, searchResults, wg)
	}
	go func() {
		wg.Wait()
		close(searchResults)
	}()
	games := make([]domain.Game, 0)
	for game := range searchResults {
		games = append(games, game)
	}
	return games
}

func getAvailableGames(branchCode int, platform string, results chan domain.Game, wg *sync.WaitGroup) {
	defer wg.Done()
	client := libClient.Client{}
	games := client.FindAvailabelGames(branchCode, platform)
	for _, game := range games {
		results <- game
	}
}
