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
		log.Panic("Error loading .env file.")
		log.Panic(err)
	}
}

func main() {

	loadEnv()
	client := twitch.NewClient(os.Getenv("NICK"), "oauth:"+os.Getenv("ACCESS_TOKEN"))

	client.Join("projared")

	client.OnConnect(func() {
		fmt.Println("Connected.")
	})

  client.OnSelfJoinMessage(func(message twitch.UserJoinMessage){
    fmt.Println("Chatbot Joined!")
  })

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		fmt.Println(message.User.Name + ": " + message.Message)
	})

	err := client.Connect()
	if err != nil {
		log.Panicln(err)
	}

}
