package controllers

import (
	"errors"
	"log"
	"os"

	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
)

func (server *Server) googleAPIAuth() *maps.Client {
	key, err := server.loadGoogleKey()
	if err != nil {
		log.Println("Load API key fatal error: %s", err)
	}
	c, err := maps.NewClient(maps.WithAPIKey(key))
	if err != nil {
		log.Println("Connect client fatal error: %s", err)
	}

	return c
}

func (server *Server) loadGoogleKey() (string, error) {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Println("Error getting env, not comming through %v", err)
		return "", errors.New("maps: destination missing")
	}

	return os.Getenv("GOOGLE_MAP_API_KEY"), nil
}
