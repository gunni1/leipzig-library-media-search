package web

import (
	"testing"

	"github.com/gunni1/leipzig-library-game-stock-api/domain"
	. "github.com/stretchr/testify/assert"
)

func TestArrangeByBranch(t *testing.T) {

	t1_stadt := domain.Media{Title: "Terminator", Branch: "Stadt", IsAvailable: true}
	t2_stadt := domain.Media{Title: "Terminator 2", Branch: "Stadt", IsAvailable: false}
	t1_gohlis := domain.Media{Title: "Terminator", Branch: "Gohlis", IsAvailable: true}
	medias := []domain.Media{t1_stadt, t2_stadt, t1_gohlis}

	result := arrangeByBranch(medias)

	expected := []MediaByBranch{
		{Branch: "Stadt", Media: []domain.Media{t1_stadt, t2_stadt}},
		{Branch: "Gohlis", Media: []domain.Media{t1_gohlis}},
	}

	Equal(t, 2, len(result))
	ElementsMatch(t, result, expected)
}

func TestEncodeBranchName(t *testing.T) {
	Equal(t, 20, encodeBranch("Bibliothek Plagwitz"))
	Equal(t, 0, encodeBranch("Stadtbibliothek"))
	Equal(t, 41, encodeBranch("Bibliothek Gohlis"))
}
