package notifier

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// AvailabilityChecker calls the library web service to check if a media item is available.
type AvailabilityChecker struct {
	libraryBaseURL string
	httpClient     *http.Client
}

// NewAvailabilityChecker creates a checker that calls the given library service base URL.
func NewAvailabilityChecker(libraryBaseURL string) *AvailabilityChecker {
	return &AvailabilityChecker{
		libraryBaseURL: libraryBaseURL,
		httpClient:     &http.Client{},
	}
}

type availabilityAPIResponse struct {
	Available bool     `json:"available"`
	Branches  []string `json:"branches"`
}

// IsAvailable queries the library API and returns true if the item is currently available.
func (checker *AvailabilityChecker) IsAvailable(sub Subscription) (bool, error) {
	params := url.Values{}
	params.Set("title", sub.Title)
	params.Set("type", sub.Type)
	params.Set("platform", sub.Platform)
	endpoint := fmt.Sprintf("%s/api/availability?%s", checker.libraryBaseURL, params.Encode())

	resp, err := checker.httpClient.Get(endpoint)
	if err != nil {
		return false, fmt.Errorf("availability check failed for %q: %w", sub.Title, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("availability API returned status %d for %q", resp.StatusCode, sub.Title)
	}

	var body availabilityAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, fmt.Errorf("could not parse availability response: %w", err)
	}
	return body.Available, nil
}
