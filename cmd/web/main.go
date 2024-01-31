package main

import (
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	libClient "github.com/gunni1/leipzig-library-game-stock-api/library-le"
)

const (
	PLATFORM = "switch"
)

func indexHandler(respWriter http.ResponseWriter, request *http.Request) {
	templ := template.Must(template.ParseFiles("cmd/web/index.html"))

	templ.Execute(respWriter, nil)
}

func gameIndexHandler(respWriter http.ResponseWriter, request *http.Request) {
	log.Print("received htmx game-index")
	branch := strings.ToLower(request.PostFormValue("branch"))
	branchCode, exists := libClient.GetBranchCode(branch)
	if !exists {
		log.Printf("Requested branch: %s does not exists.", branch)
		return
	}
	client := libClient.Client{}
	games := client.FindAvailabelGames(branchCode, PLATFORM)
	data := map[string][]domain.Game{
		"Games": games,
	}
	templ := template.Must(template.ParseFiles("cmd/web/games.html"))
	templ.Execute(respWriter, data)
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/game-index/", gameIndexHandler)
	log.Fatal(http.ListenAndServe(":3000", mux))
}
