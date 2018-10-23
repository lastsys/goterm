package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/net/http2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	STATIC_PATH              = "/static/"
	SEND_CHANNEL_BUFFER_SIZE = 16
)

// Global buffer.
var buffer Buffer

// Handler for root path.
func HomeHandler(w http.ResponseWriter, r *http.Request) {

	index, err := ioutil.ReadFile("client/index.html")
	if err != nil {
		log.Println("Failed to read index.html")
		return
	}
	w.Write(index)
}

// Upgrader for websocket end-point.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler for websocket.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Initialize channel for sending events back the clients.
	sendChannel := make(chan []byte, 16)
	go SendHandler(sendChannel, conn)

	// Initial data for buffer.
	buffer.mutex.Lock()
	sendChannel <- buffer.BufferMsg()
	buffer.mutex.Unlock()

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Fatal(err)
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

// Initialize the font if needed.
func init() {
	if _, err := os.Stat("./client/font.png"); os.IsNotExist(err) {
		log.Println("Generating font.")
		GenerateFont()
	}

	for row := 0; row < HEIGHT; row++ {
		for col := 0; col < WIDTH; col++ {
			buffer.chars[row][col].Char = 32
			buffer.chars[row][col].Foreground = 1
			buffer.chars[row][col].Background = 0
			buffer.chars[row][col].Reverse = false
		}
	}
}

// Main entry point.
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/ws", WebSocketHandler)
	r.PathPrefix(STATIC_PATH).
		Handler(http.StripPrefix(STATIC_PATH, http.FileServer(http.Dir("./client/"))))

	server := http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:9000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	http2.ConfigureServer(&server, &http2.Server{})
	fmt.Println("Starting server at", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}
