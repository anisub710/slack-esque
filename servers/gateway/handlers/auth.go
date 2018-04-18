package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"
)

//TODO: define HTTP handler functions as described in the
//assignment description. Remember to use your handler context
//struct as the receiver on these functions so that you have
//access to things like the session store and user store.

//UserHandler handles requests for the "users" resource
func (ctx *Context) UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		if !strings.HasPrefix(r.Header.Get(headerContentType), contentTypeJSON) {
			http.Error(w, "Invalid media type", http.StatusUnsupportedMediaType)
			return
		}

		newUser := &users.NewUser{}
		if err := json.NewDecoder(r.Body).Decode(newUser); err != nil {
			http.Error(w, fmt.Sprintf("error decoding JSON: %v", err), http.StatusBadRequest)
			return
		}
		user, err := newUser.ToUser()
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid user: %v", err), http.StatusInternalServerError)
		}
		ctx.UserStore.Insert(user)

	}
}
