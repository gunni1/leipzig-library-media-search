package notifier

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestBuildEmailBody_containsTitle(t *testing.T) {
	sub := Subscription{Title: "Dune", Type: "movie", Email: "user@example.com"}
	subject, body := buildEmailContent(sub)
	Contains(t, subject, "Dune")
	Contains(t, body, "Dune")
	Contains(t, body, "ausleihbar")
}

func TestBuildEmailBody_containsGameType(t *testing.T) {
	sub := Subscription{Title: "Zelda", Type: "game", Platform: "switch", Email: "user@example.com"}
	_, body := buildEmailContent(sub)
	Contains(t, body, "Zelda")
	Contains(t, body, "switch")
}
