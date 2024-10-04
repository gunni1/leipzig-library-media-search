package domain

const MOVIE string = "movie"
const GAME string = "game"

type Movie struct {
	Title       string `json:"title"`
	Branch      string `json:"branch"`
	IsAvailable string `json:"isAvailable"`
}

//Platform als DVD/Bluray verwenden? -> Gleich zu behandeln, ggf vorteile bei geneuer Suche

type Game struct {
	Title       string `json:"title"`
	Branch      string `json:"branch"`
	Platform    string `json:"platform"`
	IsAvailable string `json:"isAvailable"`
}

type Media struct {
	Title       string
	Branch      string
	Platform    string
	IsAvailable bool
}
