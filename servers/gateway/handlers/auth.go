package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"
	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"
)

//TODO: define HTTP handler functions as described in the
//assignment description. Remember to use your handler context
//struct as the receiver on these functions so that you have
//access to things like the session store and user store.

//UserHandler handles requests for the "users" resource
func (ctx *Context) UserHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		checkHeaderType(w, r, contentTypeJSON)

		newUser := &users.NewUser{}
		decodeReq(w, r, newUser)
		user, err := newUser.ToUser()
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid user: %v", err), http.StatusInternalServerError)
			return
		}
		inserted, err := ctx.UserStore.Insert(user)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error inserting user: %v", err), http.StatusInternalServerError)
			return
		}

		//refactor as method??????
		stateStruct := &SessionState{}
		sessionState, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, stateStruct)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting session state: %v", err), http.StatusInternalServerError)
			return
		}
		_, err = sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, sessionState, w)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error beginning session: %v", err), http.StatusInternalServerError)
			return
		}

		respond(w, inserted, http.StatusCreated, contentTypeJSON)

	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
	}
}

//SpecificUserHandler handles requests for a specific user
func (ctx *Context) SpecificUserHandler(w http.ResponseWriter, r *http.Request) {
	//refactor as method??????
	stateStruct := &SessionState{}
	_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, stateStruct)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting session state: %v", err), http.StatusUnauthorized)
		return
	}

	passedID := path.Base(r.URL.Path)
	reqID, err := parseID(passedID, stateStruct)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error converting user ID: %v", err), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, err := ctx.UserStore.GetByID(reqID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error finding user: %v", err), http.StatusNotFound)
			return
		}
		respond(w, user, http.StatusOK, contentTypeJSON)

	case http.MethodPatch:
		if reqID != stateStruct.User.ID {
			http.Error(w, "Action not allowed", http.StatusForbidden)
			return
		}
		checkHeaderType(w, r, contentTypeJSON)
		updates := &users.Updates{}
		decodeReq(w, r, updates)
		updatedUser, err := ctx.UserStore.Update(reqID, updates)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating user: %v", err), http.StatusInternalServerError)
		}
		respond(w, updatedUser, http.StatusOK, contentTypeJSON)

	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
	}

}

//SessionHandler handles requests for the sessions resource,
//and allows clients to begin a new session using an existing user's credentials.
func (ctx *Context) SessionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		checkHeaderType(w, r, contentTypeJSON)
		// credentials := &users.Credentials{}

	}
}

//checkHeaderType checks the header for the request and returns
// http.StatusUnsupportedMediaType if it doesn't match application/json
func checkHeaderType(w http.ResponseWriter, r *http.Request, contentType string) {
	if !strings.HasPrefix(r.Header.Get(headerContentType), contentType) {
		http.Error(w, "Invalid media type", http.StatusUnsupportedMediaType)
		return
	}
}

//decodeReq decodes the body from the request and populates it to the interface
//returns http.StatusBadRequest if there is an error
func decodeReq(w http.ResponseWriter, r *http.Request, value interface{}) {
	if err := json.NewDecoder(r.Body).Decode(value); err != nil {
		http.Error(w, fmt.Sprintf("error decoding JSON: %v", err), http.StatusBadRequest)
		return
	}
}

func parseID(passedID string, stateStruct *SessionState) (int64, error) {
	switch passedID {
	case "me":
		return stateStruct.User.ID, nil

	default:
		reqID, err := strconv.ParseInt(passedID, 10, 64)
		if err != nil {
			return 0, err
		}
		return reqID, nil
	}
}
