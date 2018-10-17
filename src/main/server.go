package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"golang.org/x/net/http2"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"time"
)

const (
	STATIC_PATH = "/static/"
)

var buffer Buffer

type KeyMsg struct {
	Msg string `json:"msg"`
	Key string `json:"key"`
}

func generateCertificate() {
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, _ := rand.Int(rand.Reader, max)
	subject := pkix.Name{
		Organization:       []string{"Combine Control Systems AB"},
		OrganizationalUnit: []string{"Internal"},
		CommonName:         "Back Office",
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject:      subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	pk, _ := rsa.GenerateKey(rand.Reader, 2048)

	derBytes, _ := x509.CreateCertificate(rand.Reader, &template, &template, &pk.PublicKey, pk)
	certOut, _ := os.Create("cert.pem")
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()

	keyOut, _ := os.Create("key.pem")
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(pk)})
	keyOut.Close()
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	index, err := ioutil.ReadFile("client/index.html")
	if err != nil {
		log.Println("Failed to read index.html")
		return
	}
	w.Write(index)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		messageType, msg, err := conn.ReadMessage()
		switch messageType {
		case websocket.BinaryMessage:
			log.Println("Got a binary message.")
		case websocket.TextMessage:
			log.Println("Got a text message.")
		}
		if err != nil {
			log.Println(err)
			return
		}

		var keyMsg KeyMsg
		err = json.Unmarshal(msg, &keyMsg)
		if err != nil {
			fmt.Println(err)
		}

		buffer.PutCharAtCursor(Character{keyMsg.Key, 1, 0})

		var bytes = buffer.Encode()
		if err := conn.WriteMessage(websocket.TextMessage, bytes); err != nil {
			log.Println(err)
			return
		}
	}
}

func init() {
	GenerateFont()
}

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
	server.ListenAndServe()
}
