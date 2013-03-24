package main

import (
	"fmt"
	"pux0r3/irc"
)

func main() {
	fmt.Printf("Irc Test\n")
	client := irc.NewIrcClient()

	messages := make(chan string)
	userInput := make(chan string)

	client.Connect("irc.rizon.net:6667", messages)
	client.Nick("IrcTest")
	client.User("PuxIRC", "PuxHost", "PuxServer", "Not Bot")

	go readInput(userInput)

	// print("func func func\n")
	var msgStr string
	var inputStr string
	for {
		select {
		case msgStr = <-messages:
			fmt.Println("recv> ", msgStr)

		case inputStr = <-userInput:
			go client.ParseInput(inputStr)
			// default:
			// 	if !client.Connected() {
			// 		break
			// 	}
		}
	}
}

func readInput(inputStream chan<- string) {
	for {
		var input string
		fmt.Scanf("%s", &input)
		inputStream <- input
		fmt.Println("input> ", input)
	}
}
