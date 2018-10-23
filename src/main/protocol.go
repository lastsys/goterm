package main

import (
	"github.com/gorilla/websocket"
)

const (
	BufferData = 0x01
	KeyPress   = 0x10
)

func HandleMessage(channel chan []byte, msg []byte) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()

	switch msg[0] {
	case KeyPress:
		buffer.PutChar(Character{charIndex(rune(msg[1])), 1, 0, false}, msg[2], msg[3])
	}

	channel <- buffer.BufferMsg()
}

func SendHandler(channel chan []byte, conn *websocket.Conn) {
	for {
		bytes := <-channel
		conn.WriteMessage(websocket.BinaryMessage, bytes)
	}
}
