package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sog01/diycctv/serverrtc"
)

type BroadcastMessage struct {
	Id         uuid.UUID
	Msg        *WebSocketMessage
	SocketConn *websocket.Conn
}

type WebSocketMessage struct {
	StreamId     string `json:"streamId"`
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
	router.HandleFunc("/ws", handleWebSocket)

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

	defer conn.Close()
	connID, _ := uuid.NewUUID()
	log.Println("New WebSocket connection, with connection ID", connID.String())

	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		broadcast <- BroadcastMessage{
			Id:         connID,
			Msg:        &msg,
			SocketConn: conn,
		}
	}

	broadcast <- BroadcastMessage{
		Id: connID,
		Msg: &WebSocketMessage{
			Type: "close",
		},
	}
	log.Println("WebSocket closed")
}

func runBroadcaster(serverChannel chan serverrtc.Message) {
	for {
		msg := <-broadcast
		serverChannel <- serverrtc.Message{
			Id:         msg.Id,
			StreamId:   msg.Msg.StreamId,
			Type:       msg.Msg.Type,
			Payload:    msg.Msg.Payload,
			SocketConn: msg.SocketConn,
		}
	}
}
