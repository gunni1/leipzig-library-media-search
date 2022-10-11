package libraryle

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestAvailability(t *testing.T) {
	False(t, isAvailable("Der gewählte Titel ist in der aktuellen Zweigstelle entliehen."))
	False(t, isAvailable("Alle Exemplare des gewählten  Titels sind entliehen."))
	False(t, isAvailable("Ein Exemplar finden Sie in einer anderen Zweigstelle."))

	True(t, isAvailable("Ein Exemplar dieses Titels wurde heute zurückgebucht."))
	True(t, isAvailable("Ein oder mehrere Exemplare dieses Titels sind in der aktuellen Zweigstelle ausleihbar."))
}
