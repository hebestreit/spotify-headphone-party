package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/hebestreit/spotify-headphone-party/party"
	"github.com/hebestreit/spotify-headphone-party/redis"
	"gopkg.in/boj/redistore.v1"
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
	redisPool := redis.NewPool()
	defer redisPool.Close()

	store, err := redistore.NewRediStoreWithPool(redisPool, []byte(os.Getenv("SESSION_KEY")))
	if err != nil {
		panic(err)
	}
	defer store.Close()

	r := mux.NewRouter()

	server := party.NewServer(store)
	go server.Listen(r)

	http.ListenAndServe(":8090", r)
}
