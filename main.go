package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gunni1/leipzig-library-media-search/watchlist"
	"github.com/gunni1/leipzig-library-media-search/web"
)

func main() {
	port := flag.Int("port", 3000, "Webserver Port")
	dataDir := flag.String("data-dir", "data", "Directory for watchlist persistence")
	checkInterval := flag.Duration("check-interval", time.Hour, "How often to check availability")

	flag.Parse()

	store, err := watchlist.NewFileStore(*dataDir)
	if err != nil {
		log.Fatalf("failed to initialise watchlist store: %v", err)
	}

	fmt.Printf("listening on port: %d \n", *port)
	mux := web.InitMux(store, *notifierURL)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
}
