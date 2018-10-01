package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/context"
	"github.com/hebestreit/spotify-party/party"
	"net/http"
	"os"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	logEnv, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		logEnv = log.InfoLevel.String()
	}

	logLevel, _ := log.ParseLevel(logEnv)
	log.SetLevel(logLevel)
}

func main() {
	server := party.NewServer()
	go server.Listen()

	http.ListenAndServe(":8090", context.ClearHandler(http.DefaultServeMux))
}
