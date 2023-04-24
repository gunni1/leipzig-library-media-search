package main

import (
	"flag"
	"fmt"

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
		fmt.Printf("Searching all games for %s \n", *platformPtr)
		games = client.GetAllAvailableGamesPlatform(*platformPtr)
	} else {
		games = client.FindAvailabelGames(*branchPtr, *platformPtr)
	}

	for _, game := range games {
		fmt.Printf("%s (%s)\n", game.Title, game.Branch)
	}

}
