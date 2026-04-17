package song

import "errors"

// below are the known Song error
var (
	ErrCountMismatch error = errors.New("the count of songs not match")
)

// Song represents a song entity
type Song struct {
	Title      string
	Artists    []string
	PreviewURL string
}

// Service is a port that defines available behavior of song package
type Service interface {
	GetSongs() ([]Song, error)
}
