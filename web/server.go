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

	mux.HandleFunc("/games/", gamesIndexHandler)
	mux.HandleFunc("/movies/", movieHandler)
	mux.HandleFunc("/games-search/", gameSearchHandler)
	mux.HandleFunc("/movies-search/", movieSearchHandler)
	return mux
}

type MoviesByBranch struct {
	Branch string
	Movies []domain.Movie
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
	movies := client.FindMovies(title)

	if len(movies) == 0 {
		fmt.Fprint(respWriter, "<p>Es wurden keine Titel gefunden.</p>")
		return
	}

	availableMovies := filterAvailable(movies)

	byBranch := arrangeByBranch(availableMovies)

	data := map[string][]MoviesByBranch{
		"Branches": byBranch,
	}
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/item-list-by-branch.html"))
	templ.Execute(respWriter, data)
}

func filterAvailable(movies []domain.Movie) []domain.Movie {
	available := make([]domain.Movie, 0)
	for _, movie := range movies {
		if movie.IsAvailable {
			available = append(available, movie)
		}
	}
	return available
}

func arrangeByBranch(movies []domain.Movie) []MoviesByBranch {
	result := make([]MoviesByBranch, 0)

	byBranch := make(map[string][]domain.Movie)
	for _, movie := range movies {
		if otherMovies, branchExists := byBranch[movie.Branch]; branchExists {
			byBranch[movie.Branch] = append(otherMovies, movie)
		} else {
			byBranch[movie.Branch] = []domain.Movie{movie}
		}
	}

	for branch, mvs := range byBranch {
		result = append(result, MoviesByBranch{Branch: branch, Movies: mvs})
	}
	return result
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
		"Items": games,
	}
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/item-list.html"))
	templ.Execute(respWriter, data)
}
