package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

type room struct {
	// forward is a channel that holds the incoming messages, that should be forwarded to other channels
	forward chan []byte

	// join is a channel for new clients requesting to join the room
	join chan *client

	// leave is a channel for existing clients requesting to leave the room
	leave chan *client

	// clients holds all the current client in this room
	clients map[*client]bool
}

// newRoom creates a new room
func newRoom() *room {
	return &room{
		forward: make(chan []byte, messageBufferSize),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		// add client
		case client := <-r.join:
			fmt.Println("new client joined!")
			r.clients[client] = true

		// remove client
		case client := <-r.leave:
			delete(r.clients, client)
			// We explicitly closed the channel to notify the end of messages.
			// It helps to terminate the for loop that range over the incoming messages of the channel.
			close(client.send)

		// forward msg to all clients
		case msg := <-r.forward:
			for c := range r.clients {
				c.send <- msg
			}
		}
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: messageBufferSize}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
	}

	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}

	r.join <- client
	defer func() { r.leave <- client }()

	go client.write()
	// the read method of client is called in the main thread to block operations (keep connection alive)
	// until it's time to close it
	client.read()
}
