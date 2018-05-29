package handlers

import (
	"time"

	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"
)

//TODO: define a session state struct for this web server
//see the assignment description for the fields you should include
//remember that other packages can only see exported fields!

//SessionState represents a session state
type SessionState struct {
	BeginTime time.Time   `json:"beginTime"`
	User      *users.User `json:"user"`
}
