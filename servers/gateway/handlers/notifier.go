package handlers

import (
	"sync"

	"github.com/gorilla/websocket"
)

//Notifier handles WebSocket Notifications
type Notifier struct {
	currConnections map[int64]*websocket.Conn
	mx              sync.Mutex
}

//NewNotifier constructs a new Notifier
func NewNotifier() *Notifier {
	n := &Notifier{
		currConnections: make(map[int64]*websocket.Conn),
	}

	return n
}

//AddClient adds a new client to the Notifier
func (n *Notifier) AddClient(client *websocket.Conn, userID int64) {
	n.mx.Lock()
	defer n.mx.Unlock()
	n.currConnections[userID] = client
	go n.readLoop(client, userID)
}

func (n *Notifier) readLoop(client *websocket.Conn, userID int64) {
	for {
		if _, _, err := client.NextReader(); err != nil {
			client.Close()
			delete(n.currConnections, userID)
			break
		}
	}
}
