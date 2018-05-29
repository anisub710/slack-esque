package handlers

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

//Notifier handles WebSocket Notifications
type Notifier struct {
	currConnections map[int64][]*websocket.Conn
	mx              sync.Mutex
}

//NewNotifier constructs a new Notifier
func NewNotifier() *Notifier {
	n := &Notifier{
		currConnections: make(map[int64][]*websocket.Conn),
	}

	return n
}

//messageInfo is a struct to hold JSON objects received from the message queue
type messageInfo struct {
	MessageType string      `json:"type,omitempty"`
	Action      interface{} `json:"action,omitempty"`
	UserIDs     []int64     `json:"userIDs,omitempty"`
}

//AddClient adds a new client to the Notifier
func (n *Notifier) AddClient(client *websocket.Conn, userID int64) {
	n.mx.Lock()
	defer n.mx.Unlock()
	n.currConnections[userID] = append(n.currConnections[userID], client)
	go n.readLoop(client, userID)
}

//readLoop reads through each connection and closes any that have an error
func (n *Notifier) readLoop(client *websocket.Conn, userID int64) {
	for {
		if _, _, err := client.NextReader(); err != nil {
			client.Close()
			connIndex := findConn(n.currConnections[userID], client)
			n.currConnections[userID] = append(n.currConnections[userID][:connIndex], n.currConnections[userID][connIndex+1:]...)
			break
		}
	}
}

//ProcessMessages broadcasts messages consumed at gateway
func (n *Notifier) ProcessMessages(messages <-chan amqp.Delivery) {
	for message := range messages {
		n.mx.Lock()
		messageInfo := &messageInfo{}
		if err := json.Unmarshal(message.Body, messageInfo); err != nil {
			log.Printf("Error decoding json object: %v", err)
			return
		}
		users := messageInfo.UserIDs
		switch len(users) {
		case 0:
			n.broadcastPublic(message)
		default:
			n.broadcastPrivate(users, message)
		}
		message.Ack(false)
		n.mx.Unlock()
	}
}

//broadcastPrivate only broadcasts to WebSockets created by users in userIDs list
func (n *Notifier) broadcastPrivate(users []int64, message amqp.Delivery) {
	for _, user := range users {
		if _, ok := n.currConnections[user]; ok {
			for _, connection := range n.currConnections[user] {
				if err := connection.WriteMessage(websocket.TextMessage, message.Body); err != nil {
					connection.Close()
					connIndex := findConn(n.currConnections[user], connection)
					n.currConnections[user] = append(n.currConnections[user][:connIndex], n.currConnections[user][connIndex+1:]...)
					return
				}
			}
		}
	}
}

//broadcastPublic broadcasts to all WebSockets
func (n *Notifier) broadcastPublic(message amqp.Delivery) {
	for user, connections := range n.currConnections {
		for _, connection := range connections {
			if err := connection.WriteMessage(websocket.TextMessage, message.Body); err != nil {
				connection.Close()
				connIndex := findConn(n.currConnections[user], connection)
				n.currConnections[user] = append(n.currConnections[user][:connIndex], n.currConnections[user][connIndex+1:]...)
				return
			}
		}
	}
}

//findConn finds the index of the connection in the slice
func findConn(a []*websocket.Conn, x *websocket.Conn) int {
	for i, n := range a {
		if x == n {
			return i
		}
	}
	return -1
}
