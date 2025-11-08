package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type SafeConn struct {
	conn *websocket.Conn
	lock *sync.RWMutex
}

func (sf *SafeConn) WriteJSON(v any) error {
	sf.lock.Lock()
	err := sf.conn.WriteJSON(v)
	sf.lock.Unlock()
	return err
}

func NewSafeConn(conn *websocket.Conn) *SafeConn {
	return &SafeConn{
		conn: conn,
		lock: &sync.RWMutex{},
	}
}
