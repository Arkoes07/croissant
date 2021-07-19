package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/Arkoes07/croissant/pkg/secret"
	"github.com/Arkoes07/croissant/pkg/secret/jsonsecret"
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

	// TODO: remove this line if secret has been used
	log.Printf("secret: %+v\n", secret)
}
