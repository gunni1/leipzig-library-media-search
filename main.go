package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gunni1/leipzig-library-game-stock-api/web"
)

func main() {
	port := flag.Int("port", 8080, "Webserver Port")
	flag.Parse()
	fmt.Printf("listening on port: %d \n", *port)
	mux := web.InitMux()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), mux))
}
