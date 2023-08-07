package main

import (
	"bytes"
	"context"
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
const CHAT_HISTORY_CAPACITY = CHAT_HISTORY_LENGTH + 30

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Panic("Error loading .env file.")
		log.Panic(err)
	}
}

func setupTwitchClient(c *twitch.Client, chat *chat) {

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
	ch chan twitch.PrivateMessage
}

func newChatCompletionRequest() openai.ChatCompletionRequest {
	return openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
        Content: "You are a Twitch chat. You receive messages and summarize them in prose as if they were your thoughts. Refrain to say they are from viewers, remember these messages are supposed to be your thoughts. Summarize them in two sentences. For example: @CohhCarnage you didn't introduce yourself in the front, you are a tresspasser enemy You hit that guy over and over and he still maintained 6 hp lol gnome ran because he was prisoner cohhStare cohhLoot cohhDerp pots? Oh for sure @mogel78 probably basic story hrs no exploring did he solve the puzzle on the second floor? @Goski11 Haha yep, insanely good game though, played through it 3 times and DOS1 2 times and absolutely loved it each time Cohh do not fight them @toajish first I thought he wanted to say something about Fortran Kappa @PSfanatic What all I said it was a shitty game. Then you attack me by calling me a child LMAO I feel like I missed this area Stay mad nerds attacks mods nope sneak in @impulsecfc @impulsecfc cohhHi Welcome to the channel! cohhGV Enjoy your stay cohhL @gokubytes dansgaming isn’t half the distance of cohh but he has the same amount of time sneak around cohh, is there a reason youre ignoring the advice of the narrator? modCheck walking in that room progresses stuff @fortntiemaster_ lol look whos talking this game has a beutifull enviorment avoided cutscene zone EZ sneak around them cohhHmm its an event You can enter thru the front you can talk to them different doorway go from the front sneak around yes yes @fortntiemaster_ No, I attacked you jumping on someone for their spelling when its clear you struggle with it yourself, try to follow along @CohhCarnage please explore ooutside untill you find a monk's grave you'll thank me later I failed the skill check to open the fridge and I forgot to save YEP you just following my foodsteps i was juat here like minutes ago :D i Go to teleport its the fastest way loot hahaha Not talk to but trigger a cut scene so they are immediately hostile @cohhcarnage long way is there a bag of holding in this game? cmon stop taking the obvious bait y'all. how long have we had the internet now hidden area Poggies @PSfanatic It's Fortnite not Fortnight. It's my name dumbass Aren’t eh. Loot! Chat fight? TPFufun PopCorn PogChamp loot!!! Chat: I can't believe CohhCarnage didn't introduce himself, what a tresspasser enemy. That guy had only 6 hp left even after being hit multiple times, lol. The gnome ran away because he was a prisoner. Oh look, pots!",
			},
		},
	}

}

func (chat *chat) summarizeChat(client *openai.Client) (summary string, err error) {

	messages := new(bytes.Buffer)

  req := newChatCompletionRequest()
  ctx := context.Background()

L:
	for {
		select {
		case m := <-chat.ch:
      _, err := messages.WriteString(m.Message + "\n")
      if err != nil {
        return "", err
      }
		default:
			break L
		}
	}

  req.Messages = append(req.Messages, openai.ChatCompletionMessage{
    Role: openai.ChatMessageRoleUser,
    Content: messages.String(),
  })

	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func main() {

	loadEnv()

	var (
		twitchClient = twitch.NewClient(os.Getenv("NICK"), "oauth:"+os.Getenv("ACCESS_TOKEN"))
		oaiClient    = openai.NewClient(os.Getenv("OPEN_AI_TOKEN"))
		chat         = chat{
			ch: make(chan twitch.PrivateMessage, CHAT_HISTORY_CAPACITY),
		}
		sigs   = make(chan os.Signal, 1)
		ticker = time.NewTicker(5 * time.Second)
		done   = make(chan int, 1)
	)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	setupTwitchClient(twitchClient, &chat)
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
			if l <= CHAT_HISTORY_LENGTH / 4 {
				fmt.Printf("length of chat: %v \n", l)
				continue
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
