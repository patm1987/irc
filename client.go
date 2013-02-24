package irc

import (
	// "errors"
	"bufio"
	"fmt"
	"net"
)

type ircClient struct {
	server     string
	connection net.Conn
	connected  bool
	nick       string

	quitStream chan bool
}

func NewIrcClient() *ircClient {
	client := new(ircClient)
	client.server = ""
	client.connection = nil
	client.connected = false
	client.nick = ""
	return client
}

func (client *ircClient) Connect(server string, output chan<- string) error {
	if client.connected {
		client.Disconnect()
	}

	client.server = server
	connection, err := net.Dial("tcp", client.server)
	if err == nil {
		client.connected = true
	} else {
		client.connected = false
		return fmt.Errorf("Failed to connect to server")
	}
	client.connection = connection

	client.quitStream = make(chan bool)
	go client.read(output, client.quitStream)

	return nil
}

func (client *ircClient) Disconnect() {
	client.connection.Close()
	client.connected = false
	client.server = ""

	client.quitStream <- true
	client.quitStream = nil
}

func (client *ircClient) Nick(nick string) error {
	if !client.connected {
		return fmt.Errorf("Not yet connected")
	}
	if client.nick == "" {
		fmt.Fprintf(client.connection, "NICK %s", nick)
	} else {
		fmt.Fprintf(client.connection, ":%s NICK %s", client.nick, nick)
	}
	client.nick = nick

	return nil
}

func (client *ircClient) send(message string) {
	fmt.Fprintln(client.connection, message)
}

func (client *ircClient) read(message chan<- string, quit chan bool) error {
	reader := bufio.NewReader(client.connection)
	readChan := make(chan string)

	go func(quit chan<- bool) {
		for {
			msgStr, err := reader.ReadString('\n')
			if err != nil {
				quit <- true
				print("error: ", err)
				break
			}
			readChan <- msgStr
		}
	}(quit)

	for {
		select {
		case msgStr := <-readChan:
			message <- msgStr

		case shouldQuit := <-quit:
			if shouldQuit {
				print("quitting")
				break
			}
		}
	}

	return nil
}
