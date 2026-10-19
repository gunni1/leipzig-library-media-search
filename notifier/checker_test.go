package notifier

import (
	"testing"

	"github.com/gunni1/leipzig-library-media-search/domain"
	. "github.com/stretchr/testify/assert"
)

type MockCatalogClient struct{}

func (client MockCatalogClient) FindMovies(title string) ([]domain.Media, error) {
	return []domain.Media{
		domain.Media{Title: "matrix", Branch: "Stadtbibliothek", Platform: "dvd", IsAvailable: true},
		domain.Media{Title: "matrix", Branch: "Gohlis", Platform: "dvd", IsAvailable: false},
	}, nil
}
func (client MockCatalogClient) FindGames(title string, platform string) ([]domain.Media, error) {
	return []domain.Media{
		domain.Media{Title: "Mario", Branch: "Stadtbibliothek", Platform: "switch", IsAvailable: true},
		domain.Media{Title: "Mario", Branch: "Gohlis", Platform: "switch", IsAvailable: true},
	}, nil
}

func (client MockCatalogClient) FindAvailableGames(branchCode int, platform string) ([]domain.Game, error) {
	return []domain.Game{domain.Game{Title: "Mario Word", Branch: "Gohlis", Platform: "switch"}}, nil
}

func (client MockCatalogClient) RetrieveReturnDate(branchCode int, platform, title string) (string, error) {
	return "", nil
}

func TestCheckAvailabilityForASubscription(t *testing.T) {
	checker := AvailabilityChecker{MockCatalogClient{}}
	matrixAvailability, err := checker.IsAvailable(Subscription{Title: "matrix", Platform: "dvd", Type: "movie"})
	Nil(t, err)
	Contains(t, matrixAvailability, "Stadtbibliothek")
	termAvailability, err := checker.IsAvailable(Subscription{Title: "terminator", Platform: "dvd", Type: "movie"})
	Nil(t, err)
	Empty(t, termAvailability)
	marioAvailability, err := checker.IsAvailable(Subscription{Title: "mario", Platform: "switch", Type: "game"})
	Nil(t, err)
	Len(t, marioAvailability, 2)
	Contains(t, marioAvailability, "Gohlis", "Stadtbibliothek")
}

func TestCheckAvailabilityForSubscriptionWithBranch(t *testing.T) {
	checker := AvailabilityChecker{MockCatalogClient{}}
	matrixSearchWithbranchGohlis, err := checker.IsAvailable(Subscription{Title: "matrix", Platform: "dvd", Type: "movie", Branch: "Gohlis"})
	Nil(t, err)
	Empty(t, matrixSearchWithbranchGohlis)
	matrixSearchWithbranchStadtBib, err := checker.IsAvailable(Subscription{Title: "matrix", Platform: "dvd", Type: "movie", Branch: "Stadtbibliothek"})
	Nil(t, err)
	Contains(t, matrixSearchWithbranchStadtBib, "Stadtbibliothek")
}
