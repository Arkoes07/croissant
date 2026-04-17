package jsonsecret

import (
	"encoding/json"
	"os"

	"github.com/Arkoes07/croissant/internal/secret"
)

// service is the adapter that will implements secret port
type service struct {
	cfg Config
}

// Config store configurable value for service
type Config struct {
	FilePath string
}

// New will create a new service
func New(cfg Config) *service {
	return &service{
		cfg: cfg,
	}
}

// secretJSON represents the structure of data in a json secret file
type secretJSON struct {
	Spotify struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"spotify"`
}

// constructSecret will construct Secret object from secretJSON
func constructSecret(sJSON *secretJSON) *secret.Secret {
	return &secret.Secret{
		Spotify: secret.Spotify{
			ClientID:     sJSON.Spotify.ClientID,
			ClientSecret: sJSON.Spotify.ClientSecret,
		},
	}
}

// Parse will parse and return the data in the secret file
func (s *service) Parse() (*secret.Secret, error) {
	secretBytes, err := os.ReadFile(s.cfg.FilePath)
	if err != nil {
		return nil, err
	}

	var sJSON secretJSON
	if err = json.Unmarshal(secretBytes, &sJSON); err != nil {
		return nil, err
	}

	return constructSecret(&sJSON), nil
}
