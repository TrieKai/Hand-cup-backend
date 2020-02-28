package api

import (
	"fmt"
	"log"
	"os"

	"handCup-project-backend/api/controllers"

	"github.com/joho/godotenv"
)

var server = controllers.Server{}

func Run() {
	var err error
	err = godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not comming through %v", err)
	} else {
		fmt.Println("We are getting the env values")
	}

	// server.Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	server.Initialize(os.Getenv("AWS_DB_DRIVER"), os.Getenv("AWS_DB_USER"), os.Getenv("AWS_DB_PASSWORD"), os.Getenv("AWS_DB_PORT"), os.Getenv("AWS_DB_HOST"), os.Getenv("AWS_DB_NAME"))

	// seed.Load(server.DB)

	server.Run(":5487")
}
