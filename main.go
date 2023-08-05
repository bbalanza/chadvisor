package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	twitch "github.com/gempir/go-twitch-irc/v4"
	godotenv "github.com/joho/godotenv"
  openai "github.com/sashabaranov/go-openai"
)

const ADDRESS string = "irc.chat.twitch.tv:6667"
const CONNECTION_TYPE string = "tcp"
const CHAT_HISTORY_LENGTH = 100
const CHAT_HISTORY_CAPACITY = CHAT_HISTORY_LENGTH + 10

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Panic("Error loading .env file.")
		log.Panic(err)
	}
}

func setupClient(c *twitch.Client, chat *[]string) {

	c.OnConnect(func() {
		fmt.Println("Connected...")
	})

	c.OnSelfJoinMessage(func(message twitch.UserJoinMessage) {
		fmt.Println("Chatbot Joined!")
	})

	c.OnPrivateMessage(func(m twitch.PrivateMessage) {
		fmtMessage := "message: " + m.Message + "\n"
		*chat = append(*chat, fmtMessage)
		if len(*chat) >= CHAT_HISTORY_LENGTH {
			*chat = make([]string, 0, CHAT_HISTORY_CAPACITY)
		}
	})

}



func main() {

	loadEnv()

	var (
		twitchClient = twitch.NewClient(os.Getenv("NICK"), "oauth:"+os.Getenv("ACCESS_TOKEN"))
    openAIClient = openai.NewClient(os.Getenv("OPEN_AI_TOKEN"))
		chat   = make([]string, 0, CHAT_HISTORY_CAPACITY)
		sigs   = make(chan os.Signal, 1)
		ticker = time.NewTicker(5 * time.Second)
		done   = make(chan int, 1)
	)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	setupClient(twitchClient, &chat)
	defer func() {
		ticker.Stop()
		twitchClient.Disconnect()
		fmt.Println("Disconnected!")
	}()

	go func() {
		twitchClient.Join("cohhcarnage")
		err := twitchClient.Connect()
		if err != nil {
			log.Panicln(err)
		}
	}()

	go func() {
		for range ticker.C {
			fmt.Println(chat)
		}
	}()

	go func() {
		<-sigs
		done <- 1
	}()

	<-done
	return
}
