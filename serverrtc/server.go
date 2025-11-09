package serverrtc

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media/ivfwriter"
	wrtc "github.com/sog01/diycctv/pkg/webrtc"
	"github.com/sog01/diycctv/pkg/ws"
)

type Message struct {
	Id         uuid.UUID
	StreamId   string
	Type       string
	Payload    string
	SocketConn *ws.SafeConn
}

type ServerRTC struct {
	serverChannel   chan Message
	peerConnections *PeerConnections
}

func NewServerRTC() *ServerRTC {
	serverChannel := make(chan Message)
	return &ServerRTC{
		serverChannel:   serverChannel,
		peerConnections: NewPeerConnections(),
	}
}

func (srv *ServerRTC) GetPeerConnectionsIds() []string {
	return srv.peerConnections.GetIds()
}

func (srv *ServerRTC) GetPeerConnectionsActiveRecords() []string {
	return srv.peerConnections.GetActiveRecords()
}

func (srv *ServerRTC) RunServerListener() chan Message {
	go func() {
		for {
			msg := <-srv.serverChannel
			switch msg.Type {
			case "offer":
				srv.handleOffer(msg)
			case "stream":
				srv.handleStream(msg)
			case "answer":
				var answer webrtc.SessionDescription
				if err := json.Unmarshal([]byte(msg.Payload), &answer); err != nil {
					log.Println("Unmarshal candidate error:", err)
					continue
				}

				peerConnection, ok := srv.peerConnections.Get(msg.Id.String())
				if !ok {
					log.Println("peer connection not exists")
					continue
				}

				log.Printf("Received Answer from browser")
				if err := peerConnection.peer.SetRemoteDescription(answer); err != nil {
					log.Println("Set Remote Description error:", err)
					continue
				}
			case "candidate":
				var candidate webrtc.ICECandidateInit
				if err := json.Unmarshal([]byte(msg.Payload), &candidate); err != nil {
					log.Println("Unmarshal candidate error:", err)
					continue
				}

				peerConnection, ok := srv.peerConnections.Get(msg.Id.String())
				if !ok {
					log.Println("peer connection not exists")
					continue
				}

				log.Printf("Received ICE candidate from browser: %s", candidate.Candidate)
				if err := peerConnection.peer.AddICECandidate(candidate); err != nil {
					log.Println("AddICECandidate error:", err)
					continue
				}
			case "close":
				peerConnection, ok := srv.peerConnections.Get(msg.Id.String())
				if !ok {
					log.Println("peer connection not exists")
					continue
				}
				srv.peerConnections.Remove(msg.Id.String())
				peerConnection.peer.Close()
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
	peerConnection, err := wrtc.CreatePeer(config, msg.SocketConn, "")
	if err != nil {
		log.Println("failed to create peer: ", err)
		return
	}
	peerConnection.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		srv.handleTrack(msg.Id.String(), tr, r)
	})
	srv.peerConnections.New(msg.Id.String(), peerConnection)
	srv.sendAnswer(msg, peerConnection)
}

func (srv *ServerRTC) handleStream(msg Message) {
	log.Println("Received stream")
	peerStream, ok := srv.peerConnections.Get(msg.StreamId)
	if !ok {
		log.Println("peer stream not found: ", msg.StreamId)
		return
	}

	if peerStream.localTrack == nil {
		log.Println("local track not ready yet, camera not connected")
		return
	}

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	peerConnection, err := wrtc.CreatePeer(config, msg.SocketConn, msg.StreamId)
	if err != nil {
		log.Println("failed to create peer: ", err)
		return
	}
	peerConnection.AddTrack(peerStream.localTrack)
	peerConnection.OnNegotiationNeeded(func() {
		srv.sendOffer(msg, peerConnection)
	})
	srv.peerConnections.New(msg.Id.String(), peerConnection)
}

func (srv *ServerRTC) sendOffer(msg Message, peerConnection *webrtc.PeerConnection) {
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		log.Println("failed create offer error:", err)
		return
	}

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		log.Println("SetRemoteDescription error:", err)
		return
	}

	offerJSON, err := json.Marshal(offer)
	if err != nil {
		log.Println("failed to marshal offer:", err)
		return
	}

	if err := msg.SocketConn.WriteJSON(map[string]any{
		"type":     "offer",
		"streamId": msg.StreamId,
		"payload":  string(offerJSON),
	}); err != nil {
		log.Println("Write offer error:", err)
		return
	}
	log.Println("Sent offer")
}

func (srv *ServerRTC) sendAnswer(msg Message, peerConnection *webrtc.PeerConnection) {
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
		"type":     "answer",
		"streamId": msg.StreamId,
		"payload":  string(answerJSON),
	}); err != nil {
		log.Println("Write answer error:", err)
		return
	}
	log.Println("Sent answer")
}

func (srv *ServerRTC) handleTrack(id string, track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	log.Printf("Got %s track: %s", track.Kind(), track.Codec().MimeType)

	if track.Kind() != webrtc.RTPCodecTypeVideo {
		return
	}

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// Use the peer connection to send RTCP
			if peerConn, ok := srv.peerConnections.Get(id); ok && peerConn.peer != nil {
				if rtcpErr := peerConn.peer.WriteRTCP([]rtcp.Packet{
					&rtcp.PictureLossIndication{MediaSSRC: uint32(track.SSRC())},
				}); rtcpErr != nil {
					log.Printf("RTCP error: %v", rtcpErr)
				} else {
					log.Println("Sent keyframe request to camera")
				}
			} else {
				break
			}
		}
	}()

	os.Mkdir("./records", 0755)

	filename := fmt.Sprintf("./records/recording_%s.ivf", id)

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

	localTrack, err := webrtc.NewTrackLocalStaticRTP(
		track.Codec().RTPCodecCapability,
		"video",
		"pion",
	)
	if err != nil {
		log.Printf("Error creating Local Track: %v", err)
		file.Close()
		return
	}
	srv.peerConnections.SetLocalTrack(id, localTrack)

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

			if err := localTrack.WriteRTP(rtpPacket); err != nil {
				log.Printf("Write local error: %v", err)
				return
			}

			if count%100 == 0 {
				log.Printf("%d packets recorded", count)
			}
		}
	}()
}
