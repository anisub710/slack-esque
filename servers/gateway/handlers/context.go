package handlers

import (
	"github.com/info344-s18/challenges-ask710/servers/gateway/indexes"
	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"
	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"
)

//TODO: define a handler context struct that
//will be a receiver on any of your HTTP
//handler functions that need access to
//globals, such as the key used for signing
//and verifying SessionIDs, the session store
//and the user store

//Context is a handler context struct that
//will be a receiver for handler functions
type Context struct {
	SigningKey   string
	SessionStore sessions.Store
	UserStore    users.Store
	Trie         *indexes.Trie
	Notifier     *Notifier
}

//NewContext constructs a new Context
func NewContext(signingKey string, sessionStore sessions.Store, userStore users.Store, trie *indexes.Trie, notifier *Notifier) *Context {
	return &Context{
		SigningKey:   signingKey,
		SessionStore: sessionStore,
		UserStore:    userStore,
		Trie:         trie,
		Notifier:     notifier,
	}
}
