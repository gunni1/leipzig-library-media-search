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

	mux.HandleFunc("/games/", gamesIndexPageHandler)
	mux.HandleFunc("/movies/", moviePageHandler)
	mux.HandleFunc("/games-search/", gameIndexHandler)
	mux.HandleFunc("/movies-search/", movieSearchHandler)
	return mux
}

type MediaByBranch struct {
	Branch string
	Media  []domain.Media
}

func gamesIndexPageHandler(respWriter http.ResponseWriter, request *http.Request) {
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/games.html"))
	templ.Execute(respWriter, nil)
}

func moviePageHandler(respWriter http.ResponseWriter, request *http.Request) {
	template := template.Must(template.ParseFS(htmlTemplates, "templates/movies.html"))
	template.Execute(respWriter, nil)
}

func gameSearchHandler(respWriter http.ResponseWriter, request *http.Request) {

}

func movieSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	title := strings.ToLower(request.PostFormValue("movie-title"))
	client := libClient.Client{}
	movies := client.FindMovies(title)

	if len(movies) == 0 {
		fmt.Fprint(respWriter, "<p>Es wurden keine Titel gefunden.</p>")
		return
	}

	availableMovies := filterAvailable(movies)
	byBranch := arrangeByBranch(availableMovies)
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
