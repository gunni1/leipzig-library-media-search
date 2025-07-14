package web

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gunni1/leipzig-library-media-search/domain"
	libClient "github.com/gunni1/leipzig-library-media-search/library-le"
)

//go:embed templates
var htmlTemplates embed.FS

//go:embed static/*
var staticHtml embed.FS

const MOVIE string = "movie"
const GAME string = "game"

// Create Mux and setup routes
func InitMux() *http.ServeMux {
	mux := http.NewServeMux()
	fileSys, _ := fs.Sub(staticHtml, "static")

	mux.Handle("/", http.FileServer(http.FS(fileSys)))
	mux.HandleFunc("/games-index/", gameIndexHandler)
	mux.HandleFunc("/movies-search/", movieSearchHandler)
	mux.HandleFunc("/games-search/", gameSearchHandler)
	mux.HandleFunc("GET /return-date/{branchCode}/{platform}/{title}", returnDateHandler)
	return mux
}

type MediaByBranch struct {
	Branch string
	Media  []domain.Media
}

type MediaTemplateData struct {
	MediaType string
	Branches  []MediaByBranch
}

func gameSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "game search")
	title := strings.ToLower(request.PostFormValue("title"))
	platform := strings.ToLower(request.PostFormValue("platform"))
	showNotAvailable := strings.ToLower(request.PostFormValue("showNotAvailable")) == "true"

	client := libClient.Client{}
	games := client.FindGames(title, platform)
	if !showNotAvailable {
		games = filterAvailable(games)
	}
	renderMediaResults(games, domain.GAME, respWriter)
}

func movieSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "movie search")
	title := strings.ToLower(request.PostFormValue("movie-title"))
	showNotAvailable := strings.ToLower(request.PostFormValue("showNotAvailable")) == "true"

	client := libClient.Client{}
	movies := client.FindMovies(title)
	if !showNotAvailable {
		movies = filterAvailable(movies)
	}
	renderMediaResults(movies, domain.MOVIE, respWriter)
}

func returnDateHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "return date")
	branchCode, _ := strconv.Atoi(request.PathValue("branchCode"))
	platform := request.PathValue("platform")
	title, _ := url.QueryUnescape(request.PathValue("title"))
	client := libClient.NewClientWithSession()
	returnDate, err := client.RetrieveReturnDate(branchCode, platform, title)
	if err != nil {
		fmt.Fprint(respWriter, "unbekannt")
		return
	}
	fmt.Fprint(respWriter, returnDate)
}

func renderMediaResults(media []domain.Media, mediaType string, respWriter http.ResponseWriter) {
	if len(media) == 0 {
		fmt.Fprint(respWriter, "<p>Es wurden keine Titel gefunden.</p>")
		return
	}
	byBranch := arrangeByBranch(media)
	data := MediaTemplateData{
		Branches:  byBranch,
		MediaType: mediaType,
	}
	templ, _ := template.New("item-list-by-branch.html").Funcs(template.FuncMap{
		"encodeBranch": encodeBranch,
	}).ParseFS(htmlTemplates, "templates/item-list-by-branch.html")
	err := templ.Execute(respWriter, data)
	log.Println(err)
}

func encodeBranch(branchName string) int {
	tokens := strings.Split(branchName, " ")
	var branch string
	if len(tokens) > 1 {
		branch = tokens[1]
	} else {
		branch = tokens[0]
	}
	code, _ := libClient.GetBranchCode(branch)
	return code
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
	defer trackExecTime(time.Now(), "game index")
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

func trackExecTime(start time.Time, desc string) {
	duration := time.Since(start)
	fmt.Printf("Request %s took: %s\n", desc, duration)
}
