package web

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gunni1/leipzig-library-media-search/domain"
	libClient "github.com/gunni1/leipzig-library-media-search/library-le"
	"github.com/gunni1/leipzig-library-media-search/watchlist"
)

//go:embed templates
var htmlTemplates embed.FS

//go:embed static/*
var staticHtml embed.FS

var wlStore *watchlist.FileStore

// sessionID reads the wl_session cookie, creating and setting one if absent.
func sessionID(w http.ResponseWriter, r *http.Request) string {
	if c, err := r.Cookie("wl_session"); err == nil && c.Value != "" {
		return c.Value
	}
	b := make([]byte, 16)
	rand.Read(b)
	id := hex.EncodeToString(b)
	setCookie(w, id)
	return id
}

func setCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "wl_session",
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   60 * 60 * 24 * 365, // 1 year
	})
}

// Create Mux and setup routes
func InitMux(store *watchlist.FileStore) *http.ServeMux {
	wlStore = store
	mux := http.NewServeMux()
	fileSys, _ := fs.Sub(staticHtml, "static")

	mux.Handle("/", http.FileServer(http.FS(fileSys)))
	mux.HandleFunc("/games-index/", gameIndexHandler)
	mux.HandleFunc("/movies-search/", movieSearchHandler)
	mux.HandleFunc("/games-search/", gameSearchHandler)
	mux.HandleFunc("GET /return-date/{branchCode}/{platform}/{title}", returnDateHandler)
	mux.HandleFunc("GET /watchlist/check", watchlistCheckHandler)
	mux.HandleFunc("POST /watchlist/toggle", watchlistToggleHandler)
	mux.HandleFunc("POST /watchlist/remove", watchlistRemoveHandler)
	mux.HandleFunc("POST /watchlist/clear", watchlistClearHandler)
	mux.HandleFunc("GET /watchlist", watchlistPageHandler)
	mux.HandleFunc("GET /watchlist/share", watchlistShareHandler)
	mux.HandleFunc("GET /watchlist/join", watchlistJoinHandler)
	return mux
}

type MediaByBranch struct {
	Branch string
	Media  []domain.Media
}

type MediaTemplateData struct {
	MediaType string
	Branches  []MediaByBranch
	Starred   map[string]bool // title -> starred for current session
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
	renderMediaResults(games, domain.GAME, respWriter, request)
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
	renderMediaResults(movies, domain.MOVIE, respWriter, request)
}

func returnDateHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "return date")
	branchCode, branchErr := strconv.Atoi(request.PathValue("branchCode"))
	if branchErr != nil {
		http.Error(respWriter, "invalid branch code", http.StatusBadRequest)
		return
	}
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

func watchlistCheckHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "watchlist check")
	title := request.URL.Query().Get("title")
	platform := request.URL.Query().Get("platform")
	mediaType := request.URL.Query().Get("type")

	client := libClient.Client{}
	var medias []domain.Media
	if mediaType == domain.MOVIE {
		medias = client.FindMovies(title)
	} else {
		medias = client.FindGames(title, platform)
	}

	// Filter to exact title match (library search is fuzzy)
	titleLower := strings.ToLower(title)
	filtered := make([]domain.Media, 0)
	for _, m := range medias {
		if strings.ToLower(m.Title) == titleLower {
			filtered = append(filtered, m)
		}
	}

	byBranch := arrangeByBranch(filtered)
	data := MediaTemplateData{
		Branches:  byBranch,
		MediaType: mediaType,
	}
	templ, err := template.New("watchlist-check.html").Funcs(template.FuncMap{
		"encodeBranch": encodeBranch,
	}).ParseFS(htmlTemplates, "templates/watchlist-check.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err := templ.Execute(respWriter, data); err != nil {
		log.Println(err)
	}
}

func renderMediaResults(media []domain.Media, mediaType string, respWriter http.ResponseWriter, request *http.Request) {
	if len(media) == 0 {
		fmt.Fprint(respWriter, "<p>Es wurden keine Titel gefunden.</p>")
		return
	}
	sid := sessionID(respWriter, request)
	starredItems := wlStore.GetAll(sid)
	starred := make(map[string]bool, len(starredItems))
	for _, item := range starredItems {
		if item.Type == mediaType {
			starred[item.Title] = true
		}
	}
	byBranch := arrangeByBranch(media)
	data := MediaTemplateData{
		Branches:  byBranch,
		MediaType: mediaType,
		Starred:   starred,
	}
	templ, err := template.New("item-list-by-branch.html").Funcs(template.FuncMap{
		"encodeBranch": encodeBranch,
	}).ParseFS(htmlTemplates, "templates/item-list-by-branch.html")
	if err != nil {
		log.Printf("template parse error: %v", err)
		http.Error(respWriter, "template error", http.StatusInternalServerError)
		return
	}
	if err := templ.Execute(respWriter, data); err != nil {
		log.Printf("template execute error: %v", err)
	}
}

func watchlistToggleHandler(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	title := r.PostFormValue("title")
	platform := r.PostFormValue("platform")
	mediaType := r.PostFormValue("type")
	item := watchlist.Item{Title: title, Platform: platform, Type: mediaType}
	starred := wlStore.Toggle(sid, item)
	data := struct {
		Title    string
		Platform string
		Type     string
		Starred  bool
	}{title, platform, mediaType, starred}
	templ := template.Must(template.ParseFS(htmlTemplates, "templates/star-button.html"))
	templ.Execute(w, data)
}

func watchlistRemoveHandler(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	title := r.PostFormValue("title")
	mediaType := r.PostFormValue("type")
	wlStore.Remove(sid, title, mediaType)
	// respond with empty body so HTMX deletes the element
}

func watchlistClearHandler(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	wlStore.Clear(sid)
	http.Redirect(w, r, "/watchlist", http.StatusSeeOther)
}

func watchlistPageHandler(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	items := wlStore.GetAll(sid)
	templ, err := template.New("watchlist-page.html").Funcs(template.FuncMap{
		"encodeBranch": encodeBranch,
	}).ParseFS(htmlTemplates, "templates/watchlist-page.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err := templ.Execute(w, items); err != nil {
		log.Println(err)
	}
}

func watchlistShareHandler(w http.ResponseWriter, r *http.Request) {
	sid := sessionID(w, r)
	scheme := extractScheme(r)
	joinURL := fmt.Sprintf("%s://%s/watchlist/join?token=%s", scheme, r.Host, sid)
	templ, err := template.ParseFS(htmlTemplates, "templates/watchlist-share.html")
	if err != nil {
		log.Println(err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	if err := templ.Execute(w, joinURL); err != nil {
		log.Println(err)
	}
}

func extractScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func watchlistJoinHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if !isValidJoinToken(token) {
		http.Error(w, "Invalid join token", http.StatusBadRequest)
		return
	}
	setCookie(w, token)
	http.Redirect(w, r, "/watchlist", http.StatusSeeOther)
}

// Join token should be a 32 character hex string (session ID)
func isValidJoinToken(token string) bool {
	tokenRE := regexp.MustCompile("^[a-fA-F0-9]{32}$")
	return tokenRE.MatchString(token)
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
