package main

import (
	"flag"
	"fmt"

	libraryleclient "github.com/gunni1/game-index-library-le/pkg/library-le"
)

func main() {
	branchPtr := flag.Int("branch", 20, "Branch code of the library")
	consolePtr := flag.String("console", "Nintendo Switch", "Console platform to list games")

	flag.Parse()

	client := libraryleclient.Client{}
	games := client.FindAvailabelGames(*branchPtr, *consolePtr)

	for _, game := range games {
		fmt.Println(game.Title)
	}

}
