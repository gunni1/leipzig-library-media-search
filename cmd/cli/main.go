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
	client := libClient.Client{}
	var media []domain.Media
	if *searchGame {
		//TODO: validate platform is set
		media = client.FindGames(*titlePtr, *platformPtr)
	}
	if *searchMovie {
		media = client.FindMovies(*titlePtr)
	}

	for _, result := range media {
		fmt.Printf("%#v\n", result)
	}

}
