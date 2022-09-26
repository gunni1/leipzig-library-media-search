package libraryleclient

type Client struct {
	baseUrl       string
	jSessionId    string
	userSessionId string
}

func (client Client) findAvailabelGames(branch int, console string) []Game {

}
