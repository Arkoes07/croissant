package service

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

// Parse will parse and return the data in the secret file
func (s *service) Parse() (*secret.Secret, error) {
	var result secret.Secret

	// open secret file
	secretFile, err := os.Open(s.cfg.FilePath)
	if err != nil {
		return &result, err
	}
	defer secretFile.Close()

	// get bytes from secret file
	secretBytes, err := ioutil.ReadAll(secretFile)
	if err != nil {
		return &result, err
	}

	// parse json data
	err = json.Unmarshal(secretBytes, &result)
	if err != nil {
		return &result, err
	}

	return &result, nil
}
