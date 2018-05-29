package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"
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
