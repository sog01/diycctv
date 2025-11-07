package serverrtc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	wrtc "github.com/sog01/diycctv/pkg/webrtc"
)

type Message struct {
	Id         uuid.UUID
	Type       string
	Payload    string
	SocketConn *websocket.Conn
}

type ServerRTC struct {
	serverChannel  chan Message
	peerConnection map[string]*webrtc.PeerConnection
	lock           *sync.RWMutex
}

func NewServerRTC() *ServerRTC {
	serverChannel := make(chan Message)
	return &ServerRTC{
		serverChannel:  serverChannel,
		peerConnection: make(map[string]*webrtc.PeerConnection),
		lock:           &sync.RWMutex{},
	}
}

func (srv *ServerRTC) RunServerListener() chan Message {
	go func() {
		for {
			msg := <-srv.serverChannel
			switch msg.Type {
			case "offer":
				srv.handleOffer(msg)
			case "candidate":
				var candidate webrtc.ICECandidateInit
				if err := json.Unmarshal([]byte(msg.Payload), &candidate); err != nil {
					log.Println("Unmarshal candidate error:", err)
					continue
				}

				srv.lock.RLock()
				peerConnection, ok := srv.peerConnection[msg.Id.String()]
				if !ok {
					log.Println("peer connection not exists")
					continue
				}
				srv.lock.RUnlock()
				log.Printf("Received ICE candidate from browser: %s", candidate.Candidate)
				if err := peerConnection.AddICECandidate(candidate); err != nil {
					log.Println("AddICECandidate error:", err)
					continue
				}
			case "close":
				srv.lock.RLock()
				peerConnection, ok := srv.peerConnection[msg.Id.String()]
				if !ok {
					log.Println("peer connection not exists")
					continue
				}
				srv.lock.RUnlock()
				peerConnection.Close()
			}
		}
	}()

	return srv.serverChannel
}

func (srv *ServerRTC) handleOffer(msg Message) {
	log.Println("Received offer")
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err := wrtc.CreatePeer(config, msg.SocketConn)
	if err != nil {
		log.Println("failed to create peer: ", err)
		return
	}
	peerConnection.OnTrack(srv.handleTrack)
	srv.lock.Lock()
	srv.peerConnection[msg.Id.String()] = peerConnection
	srv.lock.Unlock()

	var offer webrtc.SessionDescription
	if err := json.Unmarshal([]byte(msg.Payload), &offer); err != nil {
		log.Println("Unmarshal offer error:", err)
		return
	}

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		log.Println("SetRemoteDescription error:", err)
		return
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("CreateAnswer error:", err)
		return
	}

	if err := peerConnection.SetLocalDescription(answer); err != nil {
		log.Println("SetLocalDescription error:", err)
		return
	}

	answerJSON, err := json.Marshal(answer)
	if err != nil {
		log.Println("Marshal answer error:", err)
		return
	}

	if err := msg.SocketConn.WriteJSON(map[string]any{
		"type":    "answer",
		"payload": string(answerJSON),
	}); err != nil {
		log.Println("Write answer error:", err)
		return
	}
	log.Println("Sent answer")
}

func (srv ServerRTC) handleTrack(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	log.Printf("Got %s track: %s", track.Kind(), track.Codec().MimeType)

	if track.Kind() != webrtc.RTPCodecTypeVideo {
		return
	}

	os.Mkdir("./records", 0755)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("./records/recording_%s.ivf", timestamp)

	file, err := os.Create(filename)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		return
	}

	log.Printf("Recording to: %s", filename)

	ivfWriter, err := ivfwriter.NewWith(file)
	if err != nil {
		log.Printf("Error creating IVF writer: %v", err)
		file.Close()
		return
	}

	go func() {
		defer func() {
			file.Close()
			log.Printf("Recording saved: %s", filename)
		}()

		count := 0

		for {
			rtpPacket, _, readErr := track.ReadRTP()
			if readErr != nil {
				log.Printf("Track ended: %v", readErr)
				return
			}

			count++
			if count == 1 {
				log.Println("FIRST VIDEO PACKET! Recording started...")
			}

			if writeErr := ivfWriter.WriteRTP(rtpPacket); writeErr != nil {
				log.Printf("Write error: %v", writeErr)
				return
			}

			if count%100 == 0 {
				log.Printf("%d packets recorded", count)
			}
		}
	}()
}
