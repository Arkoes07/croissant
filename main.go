package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Arkoes07/croissant/pkg/secret"
	"github.com/Arkoes07/croissant/pkg/secret/jsonsecret"
	"github.com/Arkoes07/croissant/pkg/song"
	"github.com/Arkoes07/croissant/pkg/song/spotifyservice"
)

func main() {
	// generate path to directories
	var secretDir string
	{
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalln(err)
		}
		secretDir = filepath.Join(dir, "files", "secret")
	}

	// initialize secret service
	var secretSvc secret.Service
	{
		cfg := jsonsecret.Config{
			FilePath: filepath.Join(secretDir, "secret.json"),
		}
		secretSvc = jsonsecret.New(cfg)
	}

	// parse and get data from secret file
	secret, err := secretSvc.Parse()
	if err != nil {
		log.Fatalf("[main] Fail to parse secret, err := %v\n", err)
	}

	// initialize song service
	var songSvc song.Service
	{
		// TODO: get config value from config file
		cfg := spotifyservice.Config{
			PlaylistID: "37i9dQZF1DXcBWIGoYBM5M",
			SongsCount: 10,
		}
		songSvc, err = spotifyservice.New(secret.Spotify.ClientID, secret.Spotify.ClientSecret, cfg)
		if err != nil {
			log.Fatalf("[main] Fail to init song service, err := %v\n", err)
		}
	}

	// TODO: will remove soon, only for demo purpose
	songs, err := songSvc.GetSongs()
	if err != nil {
		log.Fatalf("[main] Fail to get songs, err := %v\n", err)
	}
	for _, song := range songs {
		log.Printf("%+v\n", song)
	}
}
