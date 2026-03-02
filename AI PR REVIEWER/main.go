package main

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func main() {

	err := godotenv.Load() //Loads envs from .env file.
	if err != nil {
		log.Println("Unable to load env: ", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", WebhookHandler)

	log.Println("Server running on :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
