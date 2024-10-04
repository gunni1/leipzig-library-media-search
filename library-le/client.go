package libraryle

import (
	"errors"
	"net/http"
)

const (
	LIB_BASE_URL = "https://webopac.stadtbibliothek-leipzig.de"
)

type Client struct {
	session webOpacSession
}

type webOpacSession struct {
	jSessionId    string
	userSessionId string
}

func NewClientWithSession() Client {
	client := Client{}
	client.openSession()
	return client
}

func (client *Client) openSession() error {
	resp, err := http.Get(LIB_BASE_URL + "/webOPACClient")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var session = webOpacSession{}
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case "JSESSIONID":
			session.jSessionId = cookie.Value
		case "USERSESSIONID":
			session.userSessionId = cookie.Value
		}
	}
	client.session = session
	if client.session.jSessionId == "" || client.session.userSessionId == "" {
		return errors.New("did not receive valid session ids via cookie")
	}
	return nil
}
