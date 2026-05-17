package web

import (
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
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
var notifierBaseURL string

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
func InitMux(store *watchlist.FileStore, notifierURL string) *http.ServeMux {
	wlStore = store
	notifierBaseURL = notifierURL
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
	mux.HandleFunc("GET /api/availability", availabilityHandler)
	mux.HandleFunc("POST /watchlist/subscribe", watchlistSubscribeHandler)
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

	client := libClient.NewClientWithSession()
	games, err := client.FindGames(title, platform)
	if err != nil {
		log.Printf("gameSearchHandler: %v", err)
		fmt.Fprint(respWriter, "<p>Suche fehlgeschlagen.</p>")
		return
	}
	if !showNotAvailable {
		games = filterAvailable(games)
	}
	renderMediaResults(games, domain.GAME, respWriter, request)
}

func movieSearchHandler(respWriter http.ResponseWriter, request *http.Request) {
	defer trackExecTime(time.Now(), "movie search")
	title := strings.ToLower(request.PostFormValue("movie-title"))
	showNotAvailable := strings.ToLower(request.PostFormValue("showNotAvailable")) == "true"

	client := libClient.NewClientWithSession()
	movies, err := client.FindMovies(title)
	if err != nil {
		log.Printf("movieSearchHandler: %v", err)
		fmt.Fprint(respWriter, "<p>Suche fehlgeschlagen.</p>")
		return
	}
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
		log.Printf("returnDateHandler: %v", err)
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

	client := libClient.NewClientWithSession()
	var medias []domain.Media
	var err error
	if mediaType == domain.MOVIE {
		medias, err = client.FindMovies(title)
	} else {
		medias, err = client.FindGames(title, platform)
	}

	if err != nil {
		log.Printf("watchlistCheckHandler: %v", err)
		http.Error(respWriter, "search failed", http.StatusInternalServerError)
		return
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

type availabilityResponse struct {
	Available bool     `json:"available"`
	Branches  []string `json:"branches"`
}

func availabilityHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	platform := r.URL.Query().Get("platform")
	mediaType := r.URL.Query().Get("type")
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	client := libClient.Client{}
	var medias []domain.Media
	if mediaType == domain.MOVIE {
		medias = client.FindMovies(title)
	} else {
		medias = client.FindGames(title, platform)
	}

	titleLower := strings.ToLower(title)
	available := false
	branches := make([]string, 0)
	for _, media := range medias {
		if strings.ToLower(media.Title) == titleLower && media.IsAvailable {
			available = true
			branches = append(branches, media.Branch)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availabilityResponse{Available: available, Branches: branches})
}

func watchlistSubscribeHandler(w http.ResponseWriter, r *http.Request) {
	if notifierBaseURL == "" {
		http.Error(w, "notification service not configured", http.StatusServiceUnavailable)
		return
	}
	title := r.PostFormValue("title")
	platform := r.PostFormValue("platform")
	mediaType := r.PostFormValue("type")
	email := r.PostFormValue("email")
	if title == "" || email == "" || mediaType == "" {
		http.Error(w, "title, type and email are required", http.StatusBadRequest)
		return
	}

	payload := url.Values{}
	payload.Set("title", title)
	payload.Set("platform", platform)
	payload.Set("type", mediaType)
	payload.Set("email", email)

	resp, err := http.PostForm(notifierBaseURL+"/subscriptions", payload)
	if err != nil {
		log.Printf("notifier unreachable: %v", err)
		http.Error(w, "could not reach notification service", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		http.Error(w, "notification service rejected request", resp.StatusCode)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `<p class="text-sm text-green-600 mt-2">Benachrichtigung eingerichtet ✓</p>`)
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
		log.Printf("Requested branch: %s does not exist.", branch)
		fmt.Fprint(respWriter, "<p>Bibliothek nicht gefunden.</p>")
		return
	}
	client := libClient.NewClientWithSession()
	games, err := client.FindAvailableGames(branchCode, platform)
	if err != nil {
		log.Printf("gameIndexHandler: %v", err)
		fmt.Fprint(respWriter, "<p>Suche fehlgeschlagen.</p>")
		return
	}

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
