package main

import (
	"fmt"
	"log"
	"os"

	twitch "github.com/gempir/go-twitch-irc/v4"
	godotenv "github.com/joho/godotenv"
)

const ADDRESS string = "irc.chat.twitch.tv:6667"
const CONNECTION_TYPE string = "tcp"

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("Error loading .env file.")
		log.Panic(err)
	}
}

func main() {

	loadEnv()
  client := twitch.NewClient(os.Getenv(""))

}
