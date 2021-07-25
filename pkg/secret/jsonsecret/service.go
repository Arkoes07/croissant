package jsonsecret

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/Arkoes07/croissant/pkg/secret"
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
	s := &service{
		cfg: cfg,
	}

	return s
}

// secretJSON represents the structure of data in a json secret file
type secretJSON struct {
	Spotify struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	} `json:"spotify"`
}

// constructSecret will constroct Secret object from secretJSON
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
	var result *secret.Secret

	// open secret file
	secretFile, err := os.Open(s.cfg.FilePath)
	if err != nil {
		return result, err
	}
	defer secretFile.Close()

	// get bytes from secret file
	secretBytes, err := ioutil.ReadAll(secretFile)
	if err != nil {
		return result, err
	}

	// parse json data
	var sJSON secretJSON
	err = json.Unmarshal(secretBytes, &sJSON)
	if err != nil {
		return result, err
	}

	// convert to Secret object
	result = constructSecret(&sJSON)

	return result, nil
}
