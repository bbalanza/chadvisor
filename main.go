package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func connect() net.Conn {
	conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6697")
	if err != nil {
		panic(err)
	}

	return conn
}

func disconnect(conn net.Conn, s chan os.Signal) {
  _, err := fmt.Fprintf(conn, "%s\r\n", "QUIT Bye")
  if err != nil {
    panic(err)
  }
  fmt.Println("Disconnecting")
  s <- os.Interrupt
  return
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		os.Exit(0)
	}()
  
  conn:= connect()
  disconnect(conn, c) 
  
}
