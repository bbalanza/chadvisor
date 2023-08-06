package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
  "context"

	twitch "github.com/gempir/go-twitch-irc/v4"
	godotenv "github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

const ADDRESS string = "irc.chat.twitch.tv:6667"
const CONNECTION_TYPE string = "tcp"
const CHAT_HISTORY_LENGTH = 100
const CHAT_HISTORY_CAPACITY = CHAT_HISTORY_LENGTH + 30

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Panic("Error loading .env file.")
		log.Panic(err)
	}
}

func setupClient(c *twitch.Client, chat *chat) {

	c.OnConnect(func() {
		fmt.Println("Connected...")
	})

	c.OnSelfJoinMessage(func(message twitch.UserJoinMessage) {
		fmt.Println("Chatbot Joined!")
	})

	c.OnPrivateMessage(func(m twitch.PrivateMessage) {
		chat.ch <- m
	})

}

type chat struct {
	mBuff []twitch.PrivateMessage
	ch    chan twitch.PrivateMessage
}

func (chat *chat) summarizeChat(client * openai.Client) (summary string, err error) {
	messages := make([]string, 0, CHAT_HISTORY_CAPACITY)
	for _, v := range chat.mBuff {
		messages = append(messages, v.Message)
	}

	req := openai.CompletionRequest{
		Model:       openai.GPT3TextDavinci003,
		Prompt:      messages,
		MaxTokens:   50,
		Temperature: 0,
	}

  resp, err := client.CreateCompletion(context.Background(), req)
  if err != nil {
    return "", err
  }

  return resp.Choices[0].Text, nil

}

func main() {

	loadEnv()

	var (
		twitchClient = twitch.NewClient(os.Getenv("NICK"), "oauth:"+os.Getenv("ACCESS_TOKEN"))
		oaiClient = openai.NewClient(os.Getenv("OPEN_AI_TOKEN"))
		chat = chat{
			mBuff: make([]twitch.PrivateMessage, 0, CHAT_HISTORY_CAPACITY),
			ch:    make(chan twitch.PrivateMessage, CHAT_HISTORY_CAPACITY),
		}
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
		twitchClient.Join("projared")
		err := twitchClient.Connect()
		if err != nil {
			log.Panicln(err)
		}
	}()

	go func() {
		for range ticker.C {
			l := len(chat.ch)
			if l <= 5 {
        fmt.Printf("length of chat: %v \n", l)
				continue
			}

			for i := 0; i < l; i++ {
				select {
				case m := <-chat.ch:
					chat.mBuff = append(chat.mBuff, m)
				default:
				}
			}
      
      summary, err := chat.summarizeChat(oaiClient)
      if err != nil {
        log.Printf("Error: %v", err)
        continue
      }
      fmt.Println(summary)
		}
	}()

	go func() {
		<-sigs
		done <- 1
	}()

	<-done
	return
}
