package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gunni1/leipzig-library-media-search/watchlist"
	"github.com/gunni1/leipzig-library-media-search/web"
)

func main() {
	port := flag.Int("port", 3000, "Webserver Port")
	dataDir := flag.String("data-dir", "data", "Directory for watchlist persistence")
	flag.Parse()

	store, err := watchlist.NewFileStore(*dataDir)
	if err != nil {
		log.Fatalf("failed to initialise watchlist store: %v", err)
	}

	fmt.Printf("listening on port: %d \n", *port)
	mux := web.InitMux(store)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
}
