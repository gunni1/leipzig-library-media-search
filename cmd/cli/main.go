package main

import (
	"flag"
)

func main() {
	branchPtr := flag.Int("branch", 20, "Branch code of the library")
	consolePtr := flag.String("console", "Nintendo Switch", "Console platform to list games")

	flag.Parse()

}
