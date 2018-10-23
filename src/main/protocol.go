package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

const (
	BufferData = 0x01
	KeyPress   = 0x10
)

func HandleMessage(channel chan []byte, msg []byte) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	fmt.Println(msg)

	switch msg[0] {
	case KeyPress:
		buffer.PutChar(Character{charIndex(rune(msg[1])),
			msg[4], msg[5], false}, msg[2], msg[3])
	}

	channel <- buffer.BufferMsg()
}

func SendHandler(channel chan []byte) {
	for {
		bytes := <-channel

		// Dispatch to all connected clients.
		websockets.mutex.Lock()
		for conn := range websockets.sockets {
			conn.WriteMessage(websocket.BinaryMessage, bytes)
		}
		websockets.mutex.Unlock()
	}
}

func ReadHandler(sendChannel chan []byte, conn *websocket.Conn) {
	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			websockets.Unregister(conn)
			log.Printf("We now have %v clients connected.", len(websockets.sockets))
			return
		}

		switch messageType {
		case websocket.BinaryMessage:
			go HandleMessage(sendChannel, msg)
		case websocket.TextMessage:
			log.Println("Received unexpected Text Message.")
		}
	}
}
