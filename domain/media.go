package domain

type Media struct {
	Title       string `json:"title"`
	Branch      string `json:"branch"`
	IsAvailable bool   `json:"isAvailable"`
}
