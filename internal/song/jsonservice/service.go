package jsonservice

import (
	"encoding/json"
	"math/rand"
	"os"
	"time"

	"github.com/Arkoes07/croissant/internal/song"
)

// service is the adapter that reads songs from a local JSON file.
type service struct {
	cfg Config
	rng *rand.Rand
}

// Config stores configurable values for the JSON song service.
type Config struct {
	// FilePath is the path to the JSON file containing song data.
	FilePath string
	// SongsCount is how many songs to return per GetSongs call.
	SongsCount int
}

// songJSON mirrors the JSON structure in the songs file.
type songJSON struct {
	Title      string   `json:"title"`
	Artists    []string `json:"artists"`
	PreviewURL string   `json:"preview_url"`
}

// New creates a new JSON-backed song service.
func New(cfg Config) song.Service {
	return &service{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetSongs reads all songs from the JSON file, shuffles them, and returns
// SongsCount songs. Only songs with a non-empty preview_url are included.
func (s *service) GetSongs() ([]song.Song, error) {
	data, err := os.ReadFile(s.cfg.FilePath)
	if err != nil {
		return nil, err
	}

	var raw []songJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	// Filter to songs that have a preview URL.
	var pool []song.Song
	for _, r := range raw {
		if r.PreviewURL == "" {
			continue
		}
		pool = append(pool, song.Song{
			Title:      r.Title,
			Artists:    r.Artists,
			PreviewURL: r.PreviewURL,
		})
	}

	// Shuffle so each quiz gets a different set.
	s.rng.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})

	if len(pool) <= s.cfg.SongsCount {
		return pool, nil
	}
	return pool[:s.cfg.SongsCount], nil
}
