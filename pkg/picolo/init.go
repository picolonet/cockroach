package picolo

import (
	"context"
	"firebase.google.com/go"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	"os"
	"path/filepath"
)

const SERVICE_CREDS_FILE_ENV = "SERVICE_CREDS_FILE"

var DataDir = ".picolo"

var FB_APP *firebase.App

func InitAppWithServiceAccount() *firebase.App {
	data, ok := os.LookupEnv(SERVICE_CREDS_FILE_ENV)
	if !ok {
		log.Fatal("SERVICE_CREDS_FILE_ENV is not set")
	}
	opt := option.WithCredentialsFile(data)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Infof("Error initializing app: %v", err)
		return nil
	}
	FB_APP = app
	return app
}

func CreateDataDir() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("Error getting user's home dir %v", err)
	}
	DataDir = filepath.Join(home, DataDir)
	if _, err := os.Stat(DataDir); os.IsNotExist(err) {
		if err := os.Mkdir(DataDir, 0755); err != nil {
			log.Fatalf("Error creating data store dir %v", err)
		}
	}
}
