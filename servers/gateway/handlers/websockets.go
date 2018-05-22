package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"
	"github.com/streadway/amqp"
)

//TODO: add a handler that upgrades clients to a WebSocket connection
//and adds that to a list of WebSockets to notify when events are
//read from the RabbitMQ server. Remember to synchronize changes
//to this list, as handlers are called concurrently from multiple
//goroutines.

//WebSocketHandler is a struct for upgrade requests
type WebSocketHandler struct {
	ctx      *Context
	upgrader *websocket.Upgrader
}

type messageInfo struct {
	messageType string
	action      interface{}
	userIDs     []int64
}

//NewWebSocketHandler constructs a new WebSocketsHandler
func NewWebSocketHandler(context *Context) *WebSocketHandler {
	//create, initialize, and return a new WebSocketsHandler
	return &WebSocketHandler{
		ctx: context,
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
	}
}

//ServeHTTP implements the http.Handler interface for the WebSocketsHandler
//TODO: add websocket to context and upgrade connection
//TODO: add the handler to gateway under resource path /v1/ws
func (wsh *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	stateStruct := &SessionState{}
	_, err := sessions.GetState(r, wsh.ctx.SigningKey, wsh.ctx.SessionStore, stateStruct)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting session state: %v", err), http.StatusUnauthorized)
		return
	}
	// add websocket to context
	conn, err := wsh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	//starts go routine for each connection to read incoming control messages
	wsh.ctx.Notifier.AddClient(conn, stateStruct.User.ID)

}

//TODO: start a goroutine that connects to the RabbitMQ server,
//reads events off the queue, and broadcasts them to all of
//the existing WebSocket connections that should hear about
//that event. If you get an error writing to the WebSocket,
//just close it and remove it from the list
//(client went away without closing from
//their end). Also make sure you start a read pump that
//reads incoming control messages, as described in the
//Gorilla WebSocket API documentation:
//http://godoc.org/github.com/gorilla/websocket

//TODO: check error handling
//connecToMQ connects to RabbitMQ and starts go routine to read new
//events from the queue
func (wsh *WebSocketHandler) connectToMQ(addr string, name string) error {
	conn, err := amqp.Dial("amqp://" + addr)
	if err != nil {
		return fmt.Errorf("error dialing MQ: %v", err)
	}
	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("error getting channel: %v", err)
	}
	q, err := channel.QueueDeclare(name,
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		return fmt.Errorf("error declaring queue: %v", err)
	}

	messages, err := channel.Consume(q.Name,
		"",
		false,
		false,
		false,
		false,
		nil)

	if err != nil {
		return fmt.Errorf("error consuming messages: %v", err)
	}
	go wsh.processMessages(messages)
	return nil
}

func (wsh *WebSocketHandler) processMessages(messages <-chan amqp.Delivery) error {
	for message := range messages {
		wsh.ctx.Notifier.mx.Lock()
		messageInfo := &messageInfo{}
		if err := json.Unmarshal(message.Body, messageInfo); err != nil {
			return fmt.Errorf("Error decoding json object: %v", err)
		}
		users := messageInfo.userIDs
		for k, connection := range wsh.ctx.Notifier.currConnections {
			switch {
			case len(users) == 0, len(users) != 0 && contains(users, k):
				if err := connection.WriteMessage(websocket.TextMessage, message.Body); err != nil {
					delete(wsh.ctx.Notifier.currConnections, k)
					return fmt.Errorf("Error writing message: %v", err)
				}
			}
			wsh.ctx.Notifier.mx.Unlock()
		}

		message.Ack(false)
	}
	return nil
}

func contains(userIDs []int64, id int64) bool {
	ids := make(map[int64]struct{}, len(userIDs))
	for _, i := range userIDs {
		ids[i] = struct{}{}
	}
	_, ok := ids[id]
	return ok
}

//go routine for connection. general arrangement of websockets
//
