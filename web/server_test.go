package web

import (
	"testing"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	. "github.com/stretchr/testify/assert"
)

func TestArrangeByBranch(t *testing.T) {

	t1_stadt := domain.Movie{Title: "Terminator", Branch: "Stadt", IsAvailable: true}
	t2_stadt := domain.Movie{Title: "Terminator 2", Branch: "Stadt", IsAvailable: false}
	t1_gohlis := domain.Movie{Title: "Terminator", Branch: "Gohlis", IsAvailable: true}
	movies := []domain.Movie{t1_stadt, t2_stadt, t1_gohlis}

	result := arrangeByBranch(movies)

	expected := []MoviesByBranch{
		{Branch: "Stadt", Movies: []domain.Movie{t1_stadt, t2_stadt}},
		{Branch: "Gohlis", Movies: []domain.Movie{t1_gohlis}},
	}

	Equal(t, 2, len(result))
	ElementsMatch(t, result, expected)
}
