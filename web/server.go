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

	mux.Handle("/", http.FileServer(http.Dir("web/static")))
	mux.HandleFunc("/games-index/", gameIndexHandler)
	mux.HandleFunc("/movies-search/", movieSearchHandler)
	mux.HandleFunc("/games-search/", gameSearchHandler)
	return mux
}

type MediaByBranch struct {
	Branch string
	Media  []domain.Media
}

func gameSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	title := strings.ToLower(request.PostFormValue("title"))
	platform := strings.ToLower(request.PostFormValue("platform"))
	client := libClient.Client{}
	games := client.FindGames(title, platform)
	renderMediaResults(games, respWriter)
}

func movieSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	title := strings.ToLower(request.PostFormValue("movie-title"))
	showNotAvailable := strings.ToLower(request.PostFormValue("showNotAvailable")) == "true"

	client := libClient.Client{}
	movies := client.FindMovies(title)
	if !showNotAvailable {
		movies = filterAvailable(movies)
	}
	renderMediaResults(movies, respWriter)
}

func renderMediaResults(media []domain.Media, respWriter http.ResponseWriter) {
	if len(media) == 0 {
		fmt.Fprint(respWriter, "<p>Es wurden keine Titel gefunden.</p>")
		return
	}
	byBranch := arrangeByBranch(media)
	data := map[string][]MediaByBranch{
		"Branches": byBranch,
	}
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/item-list-by-branch.html"))
	templ.Execute(respWriter, data)
}

func filterAvailable(medias []domain.Media) []domain.Media {
	available := make([]domain.Media, 0)
	for _, media := range medias {
		if media.IsAvailable {
			available = append(available, media)
		}
	}
	return available
}

func arrangeByBranch(medias []domain.Media) []MediaByBranch {
	result := make([]MediaByBranch, 0)

	byBranch := make(map[string][]domain.Media)
	for _, media := range medias {
		if otherMedias, branchExists := byBranch[media.Branch]; branchExists {
			byBranch[media.Branch] = append(otherMedias, media)
		} else {
			byBranch[media.Branch] = []domain.Media{media}
		}
	}

	for branch, mds := range byBranch {
		result = append(result, MediaByBranch{Branch: branch, Media: mds})
	}
	return result
}

func gameIndexHandler(respWriter http.ResponseWriter, request *http.Request) {
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
		"Items": games,
	}
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/item-list.html"))
	templ.Execute(respWriter, data)
}
