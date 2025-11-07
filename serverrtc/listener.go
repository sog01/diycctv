package serverrtc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
)

type Message struct {
	Type       string
	Payload    string
	SocketConn *websocket.Conn
}

type ServerRTC struct {
	serverChannel chan Message
}

func NewServerRTC() *ServerRTC {
	serverChannel := make(chan Message)
	return &ServerRTC{serverChannel: serverChannel}
}

func (srv ServerRTC) RunServerListener() chan Message {
	var (
		peerConnection *webrtc.PeerConnection
		err            error
	)

	go func() {
		for {
			msg := <-srv.serverChannel
			switch msg.Type {
			case "offer":
				log.Println("Received offer")

				config := webrtc.Configuration{
					ICEServers: []webrtc.ICEServer{
						{
							URLs: []string{"stun:stun.l.google.com:19302"},
						},
					},
				}

				peerConnection, err = webrtc.NewPeerConnection(config)
				if err != nil {
					log.Println("PC error:", err)
					return
				}
				defer peerConnection.Close()

				peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
					srv.onICECandidate(msg.SocketConn, candidate)
				})
				peerConnection.OnTrack(srv.onTrack)
				peerConnection.OnConnectionStateChange(srv.onConnectionStateChange)

				var offer webrtc.SessionDescription
				if err := json.Unmarshal([]byte(msg.Payload), &offer); err != nil {
					log.Println("Unmarshal offer error:", err)
					continue
				}

				if err := peerConnection.SetRemoteDescription(offer); err != nil {
					log.Println("SetRemoteDescription error:", err)
					continue
				}

				answer, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					log.Println("CreateAnswer error:", err)
					continue
				}

				if err := peerConnection.SetLocalDescription(answer); err != nil {
					log.Println("SetLocalDescription error:", err)
					continue
				}

				answerJSON, err := json.Marshal(answer)
				if err != nil {
					continue
				}

				if err := msg.SocketConn.WriteJSON(map[string]any{
					"type":    "answer",
					"payload": string(answerJSON),
				}); err != nil {
					log.Println("Write answer error:", err)
					continue
				}
				log.Println("Sent answer")
			case "candidate":
				var candidate webrtc.ICECandidateInit
				if err := json.Unmarshal([]byte(msg.Payload), &candidate); err != nil {
					log.Println("Unmarshal candidate error:", err)
					continue
				}

				log.Printf("Received ICE candidate from browser: %s", candidate.Candidate)
				if err := peerConnection.AddICECandidate(candidate); err != nil {
					log.Println("AddICECandidate error:", err)
					continue
				}
			}
		}
	}()

	return srv.serverChannel
}

func (srv ServerRTC) onICECandidate(socketConn *websocket.Conn, candidate *webrtc.ICECandidate) {
	if candidate == nil {
		return
	}

	candidateJSON, err := json.Marshal(candidate.ToJSON())
	if err != nil {
		return
	}

	if err := socketConn.WriteJSON(map[string]any{
		"type":    "candidate",
		"payload": string(candidateJSON),
	}); err != nil {
		log.Println("Write candidate error:", err)
	}

	log.Printf("Sent ICE candidate to browser: %s\n", candidate.String())
}

func (srv ServerRTC) onTrack(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
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

func (srv ServerRTC) onConnectionStateChange(state webrtc.PeerConnectionState) {
	log.Printf("🔗 Connection state: %s", state)
}
