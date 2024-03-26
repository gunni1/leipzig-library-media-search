package web

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	libClient "github.com/gunni1/leipzig-library-game-stock-api/library-le"
)

//go:embed templates
var htmlTemplates embed.FS

// Create Mux and setup routes
func InitMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /games/", gamesIndexHandler)
	mux.HandleFunc("GET /movies/", movieHandler)
	mux.HandleFunc("POST /games-search/", gameSearchHandler)
	mux.HandleFunc("POST /movies-search/", movieSearchHandler)
	return mux
}

func gamesIndexHandler(respWriter http.ResponseWriter, request *http.Request) {
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/games.html"))
	templ.Execute(respWriter, nil)
}

func movieHandler(respWriter http.ResponseWriter, request *http.Request) {
	template := template.Must(template.ParseFS(htmlTemplates, "templates/movies.html"))
	template.Execute(respWriter, nil)
}

func movieSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	title := strings.ToLower(request.PostFormValue("movie-title"))
	client := libClient.Client{}
	client.FindMovies(title)

}

func gameSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	branch := strings.ToLower(request.PostFormValue("branch"))
	platform := strings.ToLower(request.PostFormValue("platform"))
	branchCode, exists := libClient.GetBranchCode(branch)
	if !exists {
		log.Printf("Requested branch: %s does not exists.", branch)
		return
	}
	client := libClient.Client{}
	games := client.FindAvailabelGames(branchCode, platform)

	if len(games) == 0 {
		fmt.Fprint(respWriter, "<p>Es wurden keine ausleihbaren Titel gefunden.</p>")
		return
	}

	data := map[string][]domain.Game{
		"Games": games,
	}
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/games-list.html"))
	templ.Execute(respWriter, data)
}
