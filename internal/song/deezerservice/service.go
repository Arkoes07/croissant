package deezerservice

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Arkoes07/croissant/internal/song"
)

const (
	defaultSongsCount = 10
	playlistURL       = "https://api.deezer.com/playlist/%s/tracks?limit=%d"
)

// service is the adapter that fetches songs from the Deezer API.
type service struct {
	cfg    Config
	client *http.Client
}

// Config stores configurable values for the Deezer song service.
type Config struct {
	// SongsCount is how many songs to return per GetSongs call.
	SongsCount int
}

// deezerResponse is the top-level Deezer list response.
type deezerResponse struct {
	Data []deezerTrack `json:"data"`
}

// deezerTrack mirrors the Deezer track object.
type deezerTrack struct {
	Title   string       `json:"title"`
	Artist  deezerArtist `json:"artist"`
	Preview string       `json:"preview"`
}

// deezerArtist mirrors the Deezer artist object embedded in a track.
type deezerArtist struct {
	Name string `json:"name"`
}

// New creates a new Deezer-backed song service.
func New(cfg Config) song.Service {
	if cfg.SongsCount == 0 {
		cfg.SongsCount = defaultSongsCount
	}
	return &service{
		cfg:    cfg,
		client: &http.Client{},
	}
}

// GetSongs fetches SongsCount songs with preview URLs from the given Deezer
// playlist ID.
func (s *service) GetSongs(playlistID string) ([]song.Song, error) {
	// Request extra tracks to account for any entries without a preview URL.
	fetchCount := s.cfg.SongsCount * 3
	url := fmt.Sprintf(playlistURL, playlistID, fetchCount)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("deezerservice: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deezerservice: unexpected status %d", resp.StatusCode)
	}

	var result deezerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("deezerservice: failed to decode response: %w", err)
	}

	var songs []song.Song
	for _, t := range result.Data {
		if t.Preview == "" {
			continue
		}
		songs = append(songs, song.Song{
			Title:      t.Title,
			Artists:    []string{t.Artist.Name},
			PreviewURL: t.Preview,
		})
		if len(songs) == s.cfg.SongsCount {
			break
		}
	}

	if len(songs) != s.cfg.SongsCount {
		return songs, song.ErrCountMismatch
	}

	return songs, nil
}
