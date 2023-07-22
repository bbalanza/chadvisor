package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func makeConnection() net.Conn {
	conn, err := net.Dial("tcp", "irc://irc.chat.twitch.tv:6697")
	if err != nil {
		panic(err)
	}

	return conn
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("Interrupt Recieved!")
		os.Exit(0)
	}()

	for {
		fmt.Println("Running")
	}
}
