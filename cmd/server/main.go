package main

import (
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sog01/diycctv/pkg/rest"
	"github.com/sog01/diycctv/pkg/ws"
	"github.com/sog01/diycctv/serverrtc"
	"github.com/subosito/gotenv"
)

type BroadcastMessage struct {
	Id         uuid.UUID
	Msg        *WebSocketMessage
	SocketConn *ws.SafeConn
}

type WebSocketMessage struct {
	StreamId     string `json:"streamId"`
	Type         string `json:"type"`
	Payload      string `json:"payload"`
	IsLiveStream bool   `json:"isLiveStream"`
}

type ServerAPI struct {
	serverRtc *serverrtc.ServerRTC
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
	gotenv.Load()
	router := mux.NewRouter()
	router.HandleFunc("/ws", handleWebSocket)

	port := os.Getenv("SERVER_PORT")
	log.Println("Server starting on http://localhost:" + port)

	serverRtc := serverrtc.NewServerRTC()
	serverChan := serverRtc.RunServerListener()
	go runBroadcaster(serverChan)

	api := ServerAPI{serverRtc: serverRtc}
	router.HandleFunc("/api/live", api.handleGetLiveRecords).Methods("GET")

	log.Fatal(http.ListenAndServe(":"+port, router))
}

func (api *ServerAPI) handleGetLiveRecords(w http.ResponseWriter, r *http.Request) {
	resp := rest.Response{}
	resp.WriteSuccessJSON(w, r.Context(), http.StatusOK, api.serverRtc.GetPeerConnectionsActiveRecords())
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
	safeConn := ws.NewSafeConn(conn)
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
			SocketConn: safeConn,
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
