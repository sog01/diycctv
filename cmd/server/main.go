package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sog01/diycctv/serverrtc"
)

type BroadcastMessage struct {
	Msg        *WebSocketMessage
	SocketConn *websocket.Conn
	Type       string
}

type WebSocketMessage struct {
	Type         string `json:"type"`
	Payload      string `json:"payload"`
	IsLiveStream bool   `json:"isLiveStream"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
var broadcast = make(chan BroadcastMessage)
var offerMessage BroadcastMessage
var socketConns = make(map[string]*websocket.Conn)
var lock = sync.RWMutex{}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/ws/{type}", handleWebSocket)

	log.Println("Server starting on http://localhost:8080")

	serverRtc := serverrtc.NewServerRTC()
	serverChan := serverRtc.RunServerListener()
	go runBroadcaster(serverChan)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	vars := mux.Vars(r)

	lock.Lock()
	socketConns[vars["type"]] = conn
	lock.Unlock()

	defer conn.Close()

	log.Println("New WebSocket connection")

	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		broadcast <- BroadcastMessage{
			Msg:  &msg,
			Type: vars["type"],
		}
	}

	log.Println("WebSocket closed")
}

func runBroadcaster(serverChannel chan serverrtc.Message) {
	for {
		msg := <-broadcast
		lock.RLock()
		if msg.Msg.IsLiveStream {
			log.Println("Request From Live Streaming", msg.Msg.Type)
			if msg.Type == "cctv" {
				socketConns["live"].WriteJSON(msg.Msg)
			} else {
				socketConns["cctv"].WriteJSON(msg.Msg)
			}
		} else {
			log.Println("Request From CCTV", msg.Msg.Type)
			serverChannel <- serverrtc.Message{
				Type:       msg.Msg.Type,
				Payload:    msg.Msg.Payload,
				SocketConn: socketConns[msg.Type],
			}
		}
		lock.RUnlock()
	}
}
