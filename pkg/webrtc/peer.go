package webrtc

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/pion/webrtc/v3"
	"github.com/sog01/diycctv/pkg/ws"
)

func CreatePeer(config webrtc.Configuration, conn *ws.SafeConn, streamId string) (*webrtc.PeerConnection, error) {
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed create peer connection: %v", err)
	}

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		handleICECandidate(conn, candidate, streamId)
	})
	peerConnection.OnConnectionStateChange(handleConnectionStateChange)

	return peerConnection, nil
}

func handleICECandidate(socketConn *ws.SafeConn, candidate *webrtc.ICECandidate, streamId string) {
	if candidate == nil {
		return
	}

	candidateJSON, err := json.Marshal(candidate.ToJSON())
	if err != nil {
		return
	}

	if err := socketConn.WriteJSON(map[string]any{
		"type":     "candidate",
		"streamId": streamId,
		"payload":  string(candidateJSON),
	}); err != nil {
		log.Println("Write candidate error:", err)
	}

	log.Printf("Sent ICE candidate to browser: %s\n", candidate.String())
}

func handleConnectionStateChange(state webrtc.PeerConnectionState) {
	log.Printf("🔗 Connection state: %s", state)
}
