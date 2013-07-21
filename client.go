package irc

import (
	// "errors"
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
)

// a client to an irc server
type ircClient struct {
	server     string
	connection net.Conn
	connected  bool

	nick string
	user string

	// send a 'true' down this strem to quit (use Disconnect)
	quitStream chan bool

	outputStream chan<- string
}

func NewIrcClient() *ircClient {
	client := new(ircClient)
	client.server = ""
	client.connection = nil
	client.connected = false
	client.nick = ""
	client.user = ""
	return client
}

// Connects to a given server
// the server string should be a fully qualified name/port
// the output channel is a channel that will be written into when the server sends us a message
// return an error on failure
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

	client.outputStream = output

	client.quitStream = make(chan bool)
	go client.read()

	return nil
}

// if connected, this will disconnect our client
func (client *ircClient) Disconnect() {
	client.connection.Close()
	client.connection = nil
	client.connected = false
	client.server = ""

	client.nick = ""
	client.user = ""

	client.quitStream <- true
	client.quitStream = nil
}

func (client *ircClient) Connected() bool {
	return client.connected
}

// sendsnthe 'NICK' command to the server
// return an error if we're not connected
func (client *ircClient) Nick(nick string) error {
	if !client.connected {
		return fmt.Errorf("Not yet connected")
	}

	var sendStr string
	if client.nick == "" {
		sendStr = fmt.Sprintf("NICK %s", nick)
	} else {
		sendStr = fmt.Sprintf(":%s NICK %s", client.nick, nick)
	}
	client.send(sendStr)
	client.nick = nick

	return nil
}

// sends a user command to the connected server
// user the username
// hostname the hostname
// servername the serverhame
// realname the user's real name
// return an error if not connected
func (client *ircClient) User(user string, hostname string, servername string, realname string) error {
	if !client.connected {
		return fmt.Errorf("Not yet connected")
	}

	sendStr := fmt.Sprintf("USER %s %s %s :%s", user, hostname, servername, realname)
	client.send(sendStr)

	client.user = user

	return nil
}

func (client *ircClient) ParseInput(input string) {
	client.outputStream <- input
}

// sends a fully formatted message to the server
func (client *ircClient) send(message string) {
	fmt.Fprintln(client.connection, message)
	println("send> ", message)
}

// a goroutine for reading from the server
// message a string channel to write any received messages into
// quit a bool channel that will signal this goroutine to halt when true is sent
// returns an error when we fail to read
func (client *ircClient) read() error {
	reader := bufio.NewReader(client.connection)
	readChan := make(chan string)

	go func() {
		for {
			msgStr, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					client.quitStream <- true
				} else {
					println("err == io.EOF")
					client.quitStream <- true
				}
				println("error: ", err)
				println("recv> ", msgStr)
				break
			}
			readChan <- msgStr
		}
	}()

	for {
		select {
		case msgStr := <-readChan:
			client.parseServerString(msgStr)

		case shouldQuit := <-client.quitStream:
			if shouldQuit {
				println("quitting")
				break
			}
		}
	}

	return nil
}

func (client *ircClient) parseServerString(serverString string) {
	if strings.HasPrefix(serverString, "PING") {
		// pong
	} else {
		client.outputStream <- serverString
	}
}
