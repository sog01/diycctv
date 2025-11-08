package serverrtc

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type PeerConnection struct {
	peer       *webrtc.PeerConnection
	localTrack *webrtc.TrackLocalStaticRTP
}

type PeerConnections struct {
	lock            *sync.RWMutex
	peerConnections map[string]*PeerConnection
}

func (p *PeerConnections) New(id string, peer *webrtc.PeerConnection) {
	p.lock.Lock()
	p.peerConnections[id] = &PeerConnection{peer: peer}
	p.lock.Unlock()
}

func (p *PeerConnections) SetLocalTrack(id string, localTrack *webrtc.TrackLocalStaticRTP) {
	peerConnection, ok := p.Get(id)
	if ok {
		p.lock.Lock()
		peerConnection.localTrack = localTrack
		p.lock.Unlock()
	}
}

func (p *PeerConnections) Get(id string) (*PeerConnection, bool) {
	p.lock.RLock()
	peerConnection, ok := p.peerConnections[id]
	p.lock.RUnlock()
	return peerConnection, ok
}

func (p *PeerConnections) Remove(id string) {
	p.lock.Lock()
	delete(p.peerConnections, id)
	p.lock.Unlock()
}

func (p *PeerConnections) GetIds() []string {
	ids := []string{}
	p.lock.Lock()

	for id := range p.peerConnections {
		ids = append(ids, id)
	}

	p.lock.Unlock()
	return ids
}

func NewPeerConnections() *PeerConnections {
	return &PeerConnections{
		lock:            &sync.RWMutex{},
		peerConnections: make(map[string]*PeerConnection),
	}
}
