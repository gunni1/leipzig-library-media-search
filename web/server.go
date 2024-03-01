package web

import (
	"embed"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	libClient "github.com/gunni1/leipzig-library-game-stock-api/library-le"
)

//go:embed templates
var htmlTemplates embed.FS

type Server struct {
	Mux *http.ServeMux
}

func InitMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/game-index/", gameIndexHandler)
	return mux
}

func indexHandler(respWriter http.ResponseWriter, request *http.Request) {
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/index.html"))
	templ.Execute(respWriter, nil)
}

func gameIndexHandler(respWriter http.ResponseWriter, request *http.Request) {
	log.Print("received htmx game-index")
	branch := strings.ToLower(request.PostFormValue("branch"))
	platform := strings.ToLower(request.PostFormValue("platform"))
	branchCode, exists := libClient.GetBranchCode(branch)
	if !exists {
		log.Printf("Requested branch: %s does not exists.", branch)
		return
	}
	client := libClient.Client{}
	games := client.FindAvailabelGames(branchCode, platform)
	data := map[string][]domain.Game{
		"Games": games,
	}
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/games.html"))
	templ.Execute(respWriter, data)
}
