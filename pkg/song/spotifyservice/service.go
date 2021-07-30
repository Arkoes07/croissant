package spotifyservice

import (
	"context"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/Arkoes07/croissant/pkg/song"
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
	// validate config
	if cfg.PlaylistID == "" {
		cfg.PlaylistID = defaultPlaylistID
	}
	if cfg.SongsCount == 0 {
		cfg.SongsCount = defaultSongsCount
	}

	// create spotify client
	client, err := createSpotifyClient(clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	// construct service
	s := &service{
		cfg:    cfg,
		client: client,
	}

	return s, nil
}

// createSpotifyClient will connect and get spotify client
func createSpotifyClient(clientID string, clientSecret string) (*spotify.Client, error) {
	// construct config for auth
	authConfig := &clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     spotify.TokenURL,
	}

	// get access token
	accessToken, err := authConfig.Token(context.Background())
	if err != nil {
		return nil, err
	}

	// create and return client
	client := spotify.Authenticator{}.NewClient(accessToken)
	return &client, nil
}

// GetSongs will return list of songs with songsCount length
func (s *service) GetSongs() ([]song.Song, error) {
	var songs []song.Song
	var err error

	// get playlist object from client
	playlist, err := s.client.GetPlaylist(spotify.ID(s.cfg.PlaylistID))
	if err != nil {
		return songs, err
	}

	// get songs from playlist
	count := 0
	for _, track := range playlist.Tracks.Tracks {
		// skip if track doesn't have preview URL
		if track.Track.PreviewURL == "" {
			continue
		}

		// get artists from a playlist track
		var artists []string
		for _, artist := range track.Track.Artists {
			artists = append(artists, artist.Name)
		}

		// create song object
		song := song.Song{
			Title:      track.Track.Name,
			Artists:    artists,
			PreviewURL: track.Track.PreviewURL,
		}

		// append into songs array
		songs = append(songs, song)

		// get only songsCount songs
		count++
		if count == s.cfg.SongsCount {
			break
		}
	}

	// verify songs count
	if len(songs) != s.cfg.SongsCount {
		return songs, song.ErrCountMismatch
	}

	return songs, nil
}
