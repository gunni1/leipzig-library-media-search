package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gunni1/leipzig-library-media-search/domain"
	libClient "github.com/gunni1/leipzig-library-media-search/library-le"
)

func main() {
	searchGame := flag.Bool("game", false, "search for a game")
	searchMovie := flag.Bool("movie", false, "search for a movie")

	titlePtr := flag.String("title", "Terminator", "title to search for")
	platformPtr := flag.String("platform", "Nintendo Switch", "Console platform to list games")

	flag.Parse()

	if *searchGame && *searchMovie || !*searchGame && !*searchMovie {
		log.Fatal("please select either -movie OR -game search flag")
	}

	client := libClient.NewClientWithSession()
	var media []domain.Media
	var err error

	if *searchGame {
		media, err = client.FindGames(*titlePtr, *platformPtr)
	} else {
		media, err = client.FindMovies(*titlePtr)
	}

	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	if len(media) == 0 {
		fmt.Println("No results found")
		return
	}

	for _, result := range media {
		fmt.Printf("%#v\n", result)
	}
}
