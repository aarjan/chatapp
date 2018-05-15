package main

import (
	"log"

	"github.com/gorilla/websocket"
)

// client represents a single chatting user
type client struct {
	// socket is a websocket for this client
	socket *websocket.Conn

	// send is a channel on which messages are sent
	send chan []byte

	// room is the room in which the client is chatting
	room *room
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			log.Println("error reading: ", err)
			return
		}
		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("error writing: ", err)
			return
		}
	}
}
