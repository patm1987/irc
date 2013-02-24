package main

import (
	"fmt"
	"pux0r3/irc"
)

func main() {
	fmt.Printf("Irc Test\n")
	client := irc.NewIrcClient()

	messages := make(chan string)

	client.Connect("irc.rizon.net:6667", messages)
	for {
		msgStr := <-messages
		print("recv> ", msgStr)
	}
}
