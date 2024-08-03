package domain

type Media struct {
	Title       string `json:"title"`
	Branch      string `json:"branch"`
	BranchCode  string `json:"branchCode"`
	IsAvailable bool   `json:"isAvailable"`
}
