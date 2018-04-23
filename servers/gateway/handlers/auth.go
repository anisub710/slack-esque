package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"
	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"
	"golang.org/x/crypto/bcrypt"
)

//TODO: define HTTP handler functions as described in the
//assignment description. Remember to use your handler context
//struct as the receiver on these functions so that you have
//access to things like the session store and user store.

//UsersHandler handles requests for the "users" resource
func (ctx *Context) UsersHandler(w http.ResponseWriter, r *http.Request) {
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

		stateStruct := &SessionState{
			BeginTime: time.Now(),
			User:      inserted,
		}
		ctx.beginSession(stateStruct, r, w)

		respond(w, inserted, http.StatusCreated, contentTypeJSON)

	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
	}
}

//SpecificUserHandler handles requests for a specific user
func (ctx *Context) SpecificUserHandler(w http.ResponseWriter, r *http.Request) {

	stateStruct := &SessionState{}
	_ = ctx.getSessionState(stateStruct, r, w)

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
			return
		}
		respond(w, updatedUser, http.StatusOK, contentTypeJSON)

	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
	}

}

//SessionsHandler handles requests for the sessions resource,
//and allows clients to begin a new session using an existing user's credentials.
func (ctx *Context) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		checkHeaderType(w, r, contentTypeJSON)
		credentials := &users.Credentials{}
		decodeReq(w, r, credentials)

		findUser, err := ctx.UserStore.GetByEmail(credentials.Email)

		//take about the same amount of time as authenticating and then respond with a http.StatusUnauthorized
		if err != nil {
			bcrypt.CompareHashAndPassword([]byte("password"), []byte("wastetime"))
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		if err = findUser.Authenticate(credentials.Password); err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		stateStruct := &SessionState{}
		ctx.beginSession(stateStruct, r, w)

		respond(w, findUser, http.StatusCreated, contentTypeJSON)

	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
	}
}

//SpecificSessionHandler handles requests related to a specific authenticated session
func (ctx *Context) SpecificSessionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		segment := path.Base(r.URL.Path)
		if segment != "mine" {
			http.Error(w, "Forbidden User", http.StatusForbidden)
			return
		}
		_, err := sessions.EndSession(r, ctx.SigningKey, ctx.SessionStore)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error ending session: %v", err), http.StatusInternalServerError)
			return
		}
		respond(w, "Signed Out", http.StatusOK, contentTypeText)

	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
	}
}

//AvatarHandler handles requests related to changing profile pictures
func (ctx *Context) AvatarHandler(w http.ResponseWriter, r *http.Request) {
	stateStruct := &SessionState{}
	_ = ctx.getSessionState(stateStruct, r, w)
	vars := mux.Vars(r)
	passedID := vars["id"]
	reqID, err := parseID(passedID, stateStruct)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error converting user ID: %v", err), http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodPut:
		if reqID != stateStruct.User.ID {
			http.Error(w, "Action not allowed", http.StatusForbidden)
			return
		}
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("avatar")
		defer file.Close()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting image: %v", err), http.StatusForbidden)
			return
		}
		fileType := strings.Split(handler.Filename, ".")
		fileName := string(reqID) + fileType[len(fileType)-1]
		f, err := os.OpenFile("/avatars/"+fileName, os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		io.Copy(f, file)
		if _, err = ctx.UserStore.UpdatePhoto(reqID, fileName); err != nil {
			http.Error(w, fmt.Sprintf("Error updating photo: %v", err), http.StatusInternalServerError)
			return
		}
	case http.MethodGet:
		fileName := stateStruct.User.PhotoURL
		if _, err := os.Stat("/avatars/" + fileName); os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Could not find photo: %v", err), http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, fileName)

	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
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

//CHANGE
//put checkHeaderType in decodeReq and return err and statusCode
//decodeReq decodes the body from the request and populates it to the interface
//returns http.StatusBadRequest if there is an error
func decodeReq(w http.ResponseWriter, r *http.Request, value interface{}) {
	if err := json.NewDecoder(r.Body).Decode(value); err != nil {
		http.Error(w, fmt.Sprintf("error decoding JSON: %v", err), http.StatusBadRequest)
		return
	}
}

//parseID checks the UserID and converts the string to int if necessary
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

//getSessionState calls sessions.GetState
func (ctx *Context) getSessionState(stateStruct *SessionState, r *http.Request, w http.ResponseWriter) sessions.SessionID {
	sessionState, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, stateStruct)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting session state: %v", err), http.StatusInternalServerError)
		return sessions.InvalidSessionID
	}

	return sessionState
}

//remove helper
//beginSession calls sessions.BeginSession
func (ctx *Context) beginSession(stateStruct *SessionState, r *http.Request, w http.ResponseWriter) {
	_, err := sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, stateStruct, w)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error beginning session: %v", err), http.StatusInternalServerError)
		return
	}
}
