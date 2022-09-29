package main

import (
	"flag"

	libraryleclient "github.com/gunni1/game-index-library-le/pkg/library-le"
)

func main() {
	branchPtr := flag.Int("branch", 20, "Branch code of the library")
	consolePtr := flag.String("console", "Nintendo Switch", "Console platform to list games")

	flag.Parse()

	client := libraryleclient.Client{}
	client.FindAvailabelGames(*branchPtr, *consolePtr)

}
