package libraryle

import (
	"crypto/tls"
	"errors"
	"net/http"
	"time"
)

const (
	LIB_BASE_URL = "https://bibliothekskatalog.leipzig.de"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	},
}

type Client struct {
	session webOpacSession
}

type webOpacSession struct {
	jSessionId    string
	userSessionId string
}

func NewClientWithSession() Client {
	client := Client{}
	client.newSession()
	return client
}

func (client *Client) newSession() error {
	resp, err := httpClient.Get(LIB_BASE_URL + "/webOPACClient")
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
