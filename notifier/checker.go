package notifier

import (
	"strings"

	"github.com/gunni1/leipzig-library-media-search/catalog"
	"github.com/gunni1/leipzig-library-media-search/domain"
	libClient "github.com/gunni1/leipzig-library-media-search/library-le"
)

// AvailabilityChecker use a catalog client to check the availability
type AvailabilityChecker struct {
	client catalog.Client
}

// NewAvailabilityChecker creates a checker that calls the given library service base URL.
func NewAvailabilityChecker() *AvailabilityChecker {
	return &AvailabilityChecker{libClient.NewClientWithSession()}
}

type availabilityAPIResponse struct {
	Available bool     `json:"available"`
	Branches  []string `json:"branches"`
}

// IsAvailable queries the library API and returns true if the item is currently available.
func (checker *AvailabilityChecker) IsAvailable(sub Subscription) ([]string, error) {
	client := checker.client
	var medias []domain.Media
	var err error
	if sub.Type == domain.MOVIE {
		medias, err = client.FindMovies(sub.Title)
	} else {
		medias, err = client.FindGames(sub.Title, sub.Platform)
	}
	if err != nil {
		return nil, err
	}
	resultBranches := make([]string, 0)
	for _, media := range medias {
		if sub.Branch != "" && sub.Branch != media.Branch {
			continue
		}
		if strings.ToLower(media.Title) == strings.ToLower(sub.Title) && media.IsAvailable {
			resultBranches = append(resultBranches, media.Branch)
		}
	}
	return resultBranches, nil
}
