package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Arkoes07/croissant/pkg/secret"
	secretservice "github.com/Arkoes07/croissant/pkg/secret/service"
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
		cfg := secretservice.Config{
			FilePath: filepath.Join(secretDir, "secret.json"),
		}
		secretSvc = secretservice.New(cfg)
	}

	// parse and get data from secret file
	secret, err := secretSvc.Parse()
	if err != nil {
		log.Fatalf("[main] Fail to parse secret, err := %v\n", err)
	}

	// TODO: remove this line if secret has been used
	log.Printf("secret: %+v\n", secret)
}
