package spotifyservice

import (
	"context"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/Arkoes07/croissant/internal/song"
)

// below are some constants that could be used by service
const (
	defaultPlaylistID string = "37i9dQZF1DXcBWIGoYBM5M"
	defaultSongsCount int    = 10
)

// service is the adapter that will implements song port
type service struct {
	cfg    Config
	client *spotify.Client
}

// Config store configurable value for service
type Config struct {
	PlaylistID string
	SongsCount int
}

// New will create a new service
func New(clientID string, clientSecret string, cfg Config) (*service, error) {
	if cfg.PlaylistID == "" {
		cfg.PlaylistID = defaultPlaylistID
	}
	if cfg.SongsCount == 0 {
		cfg.SongsCount = defaultSongsCount
	}

	client, err := createSpotifyClient(clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	return &service{
		cfg:    cfg,
		client: client,
	}, nil
}

// createSpotifyClient will connect and return a spotify client
func createSpotifyClient(clientID string, clientSecret string) (*spotify.Client, error) {
	authConfig := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotifyauth.TokenURL,
	}

	httpClient := authConfig.Client(context.Background())
	client := spotify.New(httpClient)
	return client, nil
}

// GetSongs will return list of songs with songsCount length
func (s *service) GetSongs() ([]song.Song, error) {
	playlist, err := s.client.GetPlaylist(context.Background(), spotify.ID(s.cfg.PlaylistID))
	if err != nil {
		return nil, err
	}

	var songs []song.Song
	for _, item := range playlist.Tracks.Tracks {
		if item.Track.PreviewURL == "" {
			continue
		}

		var artists []string
		for _, artist := range item.Track.Artists {
			artists = append(artists, artist.Name)
		}

		songs = append(songs, song.Song{
			Title:      item.Track.Name,
			Artists:    artists,
			PreviewURL: item.Track.PreviewURL,
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
