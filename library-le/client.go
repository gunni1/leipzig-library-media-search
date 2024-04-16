package libraryle

import (
	"errors"
	"net/http"
)

const (
	LIB_BASE_URL = "https://webopac.stadtbibliothek-leipzig.de"
)

// Deprecated?
var BranchCodes = map[int]string{
	0:  "Stadtbibliothek",
	20: "Bibliothek Plagwitz",
	21: "Bibliothek Wiederitzsch",
	22: "Bibliothek Böhlitz-Ehrenberg",
	23: "Bibl. Lützschena-Stahmeln",
	25: "Bibliothek Holzhausen",
	30: "Bibliothek Südvorstadt",
	41: "Bibliothek Gohlis",
	50: "Bibliothek Volkmarsdorf",
	51: "Bibliothek Schönefeld",
	60: "Bibliothek Paunsdorf",
	61: "Bibliothek Reudnitz",
	70: "Bibliothek Mockau",
	82: "Bibliothek Grünau-Mitte",
	83: "Bibliothek Grünau-Nord",
	84: "Bibliothek Grünau-Süd",
	90: "Fahrbibliothek",
}

// Deprecated?
func BranchCodeKeys() []int {
	keys := make([]int, 0, len(BranchCodes))
	for key := range BranchCodes {
		keys = append(keys, key)
	}
	return keys
}

type Client struct {
	session webOpacSession
}

type webOpacSession struct {
	jSessionId    string
	userSessionId string
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
