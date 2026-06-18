package catalog

import (
	"github.com/gunni1/leipzig-library-media-search/domain"
)

// Client defines the interface for querying the library catalog.
// Implementations hide the details of HTTP session management,
// request building, and HTML parsing.
type Client interface {
	// FindMovies searches for movies by title across all library branches.
	// Returns a slice of Media or an error if the search fails.
	// Empty results (nil error, empty slice) are different from errors.
	FindMovies(title string) ([]domain.Media, error)

	// FindGames searches for games by title and platform across all library branches.
	// Returns a slice of Media or an error if the search fails.
	FindGames(title, platform string) ([]domain.Media, error)

	// FindAvailableGames lists all games available in a specific branch for a platform.
	// Returns a slice of Game or an error if the search fails.
	FindAvailableGames(branchCode int, platform string) ([]domain.Game, error)

	// RetrieveReturnDate returns the next available return date for a title in a specific branch.
	// Returns the date as a string (format: DD.MM.YYYY) or an error if not found.
	RetrieveReturnDate(branchCode int, platform, title string) (string, error)
}
