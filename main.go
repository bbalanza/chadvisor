package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	godotenv "github.com/joho/godotenv"
)

const ADDRESS string = "irc.chat.twitch.tv:6667"
const CONNECTION_TYPE string = "tcp"

type Bot struct {
	conn net.Conn
	cred Credentials
}

type Credentials struct {
	appID       string
	accessToken string
}

func (bot *Bot) Init(connType string, addr string) {
	bot.GetCredentials()
	bot.Connect(connType, addr)
	bot.SetCapabilities()
	bot.Authenticate()
}

func (bot Bot) SetCapabilities() {

	err := bot.SendMessage("CAP", "REQ", ":twitch.tv/commands twitch.tv/membership")
	if err != nil {
		log.Panicln(err)
	}
}

func (bot Bot) SendMessage(messages ...string) error {
	m := strings.Join(messages, " ")
	_, err := fmt.Fprintf(bot.conn, "%s\r\n", m)
	if err != nil {
		log.Println("Could not write to wire, check your connection plz.")
		log.Println(err)
	}
	return err
}

func (bot *Bot) Connect(connType string, addr string) {
	conn, err := net.Dial(connType, addr)
	if err != nil {
		log.Panic(err)
	}
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	bot.conn = conn
	return
}

func (bot Bot) Disconnect() {
	err := bot.conn.Close()
	if err != nil {
		log.Panicln("Could not disconnect")
		log.Panic(err)
	}
	fmt.Println("Disconnecting")
	return
}

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Panicln("Error loading .env file.")
		log.Panic(err)
	}
}

func (bot *Bot) GetCredentials() {
	loadEnv()
	bot.cred.accessToken = os.Getenv("ACCESS_TOKEN")
	bot.cred.appID = os.Getenv("APP_ID")

	if bot.cred.appID == "" {
		log.Panicf("Could not find app ID.")
	}
	if bot.cred.accessToken == "" {
		log.Panicf("Could not find access token.")
	}

	return
}

func (bot Bot) Authenticate() {

	if bot.cred == (Credentials{}) {
		log.Panicln("Bot credentials are not initialized.")
	}

	if bot.conn == nil {
		log.Panicln("Bot has no active connection.")
	}

	err := bot.SendMessage("PASS", "oauth:"+bot.cred.accessToken)
	if err != nil {
		log.Panicln("Could not send PASS message.")
		log.Panic(err)
	}

	err = bot.SendMessage("NICK", "statheros")
	if err != nil {
		log.Panicln("Could not set NICK.")
		log.Panic(err)
	}

	return

}

func (bot Bot) ReadResponses() {
	connReader := bufio.NewReader(bot.conn)
	message := ""
	go func() {
		for {
			message, _ = connReader.ReadString('\n')
			fmt.Print(message)
			time.Sleep(1 * time.Second)
		}
	}()
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	bot := Bot{}

	go func() {
		<-c
		bot.Disconnect()
		os.Exit(0)
	}()

	bot.Init(CONNECTION_TYPE, ADDRESS)
	bot.ReadResponses()

  for {
    
  }
}
