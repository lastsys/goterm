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
	"sync"
	"time"
)

const (
	STATIC_PATH              = "/static/"
	SEND_CHANNEL_BUFFER_SIZE = 256
	SOCKET_READ_BUFFER_SIZE  = 1024
	SOCKET_WRITE_BUFFER_SIZE = 1024
	WRITE_TIMEOUT            = 15
	READ_TIMEOUT             = 15
)

// Global buffer.
var buffer Buffer

// Register of websockets.
type socketRegister struct {
	// Map used as a set.
	sockets map[*websocket.Conn]bool
	mutex   sync.Mutex
}

var websockets socketRegister

func (s *socketRegister) Register(conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.sockets[conn] = true
}

func (s *socketRegister) Unregister(conn *websocket.Conn) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.sockets, conn)
}

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
	ReadBufferSize:  SOCKET_READ_BUFFER_SIZE,
	WriteBufferSize: SOCKET_WRITE_BUFFER_SIZE,
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

	log.Println("Registering connection.")
	websockets.Register(conn)
	log.Printf("We now have %v clients connected.", len(websockets.sockets))

	// Initialize channel for sending events back the clients.
	sendChannel := make(chan []byte, SEND_CHANNEL_BUFFER_SIZE)
	go SendHandler(sendChannel)

	// Initial data for buffer.
	buffer.mutex.Lock()
	sendChannel <- buffer.BufferMsg()
	buffer.mutex.Unlock()

	go ReadHandler(sendChannel, conn)
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

	// Maps need to be initialized properly.
	websockets.sockets = make(map[*websocket.Conn]bool)
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
		Addr:         "0.0.0.0:9000",
		WriteTimeout: WRITE_TIMEOUT * time.Second,
		ReadTimeout:  READ_TIMEOUT * time.Second,
	}
	http2.ConfigureServer(&server, &http2.Server{})
	fmt.Println("Starting server at", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}
