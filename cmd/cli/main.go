package main

import (
	"flag"
	"fmt"

	"github.com/gunni1/leipzig-library-media-search/domain"
	libClient "github.com/gunni1/leipzig-library-media-search/library-le"
)

func main() {
	branchPtr := flag.Int("branch", 20, "Branch code of the library")
	platformPtr := flag.String("platform", "Nintendo Switch", "Console platform to list games")

	flag.Parse()

	client := libClient.Client{}

	var games []domain.Game
	games = client.FindAvailabelGames(*branchPtr, *platformPtr)

	for _, game := range games {
		fmt.Printf("%s (%s)\n", game.Title, game.Branch)
	}

}
