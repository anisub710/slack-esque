package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"
	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"
	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/crypto/bcrypt"
)

//TODO: define HTTP handler functions as described in the
//assignment description. Remember to use your handler context
//struct as the receiver on these functions so that you have
//access to things like the session store and user store.

type resetInfo struct {
	ResetPass    string `json:"resetPass"`
	Password     string `json:"password"`
	PasswordConf string `json:"passwordConf"`
}

//UsersHandler handles requests for the "users" resource
func (ctx *Context) UsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		newUser := &users.NewUser{}
		code, err := decodeReq(w, r, newUser)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error with provided data: %v", err), code)
			return
		}

		strength := zxcvbn.PasswordStrength(newUser.Password, nil)

		if strength.Score <= 2 {
			http.Error(w, "Password is not strong enough", http.StatusBadRequest)
			return
		}

		user, err := newUser.ToUser()
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid user: %v", err), http.StatusBadRequest)
			return
		}

		inserted, err := ctx.UserStore.Insert(user)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error inserting user: %v", err), http.StatusInternalServerError)
			return
		}
		ctx.Trie.AddConvertedUsers(inserted.FirstName, inserted.LastName, inserted.UserName, inserted.ID)
		stateStruct := &SessionState{
			BeginTime: time.Now(),
			User:      inserted,
		}
		_, err = sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, stateStruct, w)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error beginning session: %v", err), http.StatusInternalServerError)
			return
		}

		respond(w, inserted, http.StatusCreated, ContentTypeJSON)

	case http.MethodGet:
		stateStruct := &SessionState{}
		_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, stateStruct)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting session state: %v", err), http.StatusInternalServerError)
			return
		}
		queries := r.URL.Query().Get("q")
		if len(queries) < 1 {
			http.Error(w, "Missing 'q' query string parameter", http.StatusBadRequest)
			return
		}
		userIDs := ctx.Trie.Find(queries, 20)
		users, err := ctx.UserStore.GetSearchUsers(userIDs)
		if users == nil || err != nil {
			http.Error(w, fmt.Sprintf("Error getting users based on search: %v", err), http.StatusBadRequest)
		}
		respond(w, users, http.StatusOK, ContentTypeJSON)
	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
	}
}

//SpecificUserHandler handles requests for a specific user
func (ctx *Context) SpecificUserHandler(w http.ResponseWriter, r *http.Request) {

	stateStruct := &SessionState{}
	_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, stateStruct)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting session state: %v", err), http.StatusUnauthorized)
		return
	}

	passedID := path.Base(r.URL.Path)
	// vars := mux.Vars(r)
	// passedID := vars["id"]
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
		respond(w, user, http.StatusOK, ContentTypeJSON)

	case http.MethodPatch:
		if reqID != stateStruct.User.ID {
			http.Error(w, "Action not allowed", http.StatusForbidden)
			return
		}
		updates := &users.Updates{}
		code, err := decodeReq(w, r, updates)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error with provided data: %v", err), code)
			return
		}
		updatedUser, err := ctx.UserStore.Update(reqID, updates)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating user: %v", err), http.StatusInternalServerError)
			return
		}
		ctx.Trie.RemoveConvertedUsers(stateStruct.User.FirstName, stateStruct.User.LastName, stateStruct.User.ID)
		ctx.Trie.AddConvertedUsers(updatedUser.FirstName, updatedUser.LastName, updatedUser.UserName, updatedUser.ID)
		respond(w, updatedUser, http.StatusOK, ContentTypeJSON)

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
		credentials := &users.Credentials{}
		code, err := decodeReq(w, r, credentials)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error with provided data: %v", err), code)
			return
		}
		findUser, err := ctx.UserStore.GetByEmail(credentials.Email)

		if err != nil {
			bcrypt.CompareHashAndPassword([]byte("password"), []byte("wastetime"))
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		ipaddr := getClientKey(r)
		currFails, err := ctx.SessionStore.Increment(ipaddr, 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("error saving failed attempts1 : %v", err), http.StatusInternalServerError)
			return
		}
		if currFails >= 5 {
			ctx.SessionStore.Increment(ipaddr, 1)
			currTimeLeft, _ := ctx.SessionStore.TimeLeft(ipaddr)
			w.Header().Add(HeaderRetryAfter, HeaderRetryAfter)
			http.Error(w, fmt.Sprintf("Too many failed attempts. Try again in %s minutes", currTimeLeft), http.StatusTooManyRequests)
			return
		}

		if err = findUser.Authenticate(credentials.Password); err != nil {
			if _, err := ctx.SessionStore.Increment(ipaddr, 1); err != nil {
				http.Error(w, fmt.Sprintf("error saving failed attempts2 : %v", err), http.StatusInternalServerError)
				return
			}
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		// add to userslogin
		login := &users.Login{
			Userid:    findUser.ID,
			LoginTime: time.Now(),
			IPAddr:    getClientKey(r),
		}

		_, err = ctx.UserStore.InsertLogin(login)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error inserting login: %v", err), http.StatusInternalServerError)
			return
		}
		stateStruct := &SessionState{
			BeginTime: time.Now(),
			User:      findUser,
		}
		_, err = sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, stateStruct, w)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error beginning session: %v", err), http.StatusInternalServerError)
			return
		}

		respond(w, findUser, http.StatusCreated, ContentTypeJSON)

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
		respond(w, "Signed Out", http.StatusOK, ContentTypeText)

	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
	}
}

//AvatarHandler handles requests related to changing profile pictures
func (ctx *Context) AvatarHandler(w http.ResponseWriter, r *http.Request) {
	stateStruct := &SessionState{}
	_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, stateStruct)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting session state: %v", err), http.StatusInternalServerError)
		return
	}
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
		strID := strconv.FormatInt(reqID, 10)
		// filePath := fmt.Sprintf("/v1/users/%s/avatar", strID)
		fileName := strID + "." + fileType[len(fileType)-1]
		f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0666)
		defer f.Close()
		if err != nil {
			http.Error(w, fmt.Sprintf("Error uploading photo: %v", err), http.StatusInternalServerError)
			return
		}

		io.Copy(f, file)
		if _, err = ctx.UserStore.UpdatePhoto(reqID, fileName); err != nil {
			http.Error(w, fmt.Sprintf("Error updating photo: %v", err), http.StatusInternalServerError)
			return
		}

		respond(w, "Image successfully uploaded", http.StatusOK, ContentTypeText)

	case http.MethodGet:
		user, err := ctx.UserStore.GetByID(reqID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting user: %v", err), http.StatusInternalServerError)
			return
		}
		fileName := user.PhotoURL
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("Could not find photo: %v", err), http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, fileName)

	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return
	}

}

//ResetHandler handles requests to reset passwords
func (ctx *Context) ResetHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		resetStruct := &users.PassReset{}
		code, err := decodeReq(w, r, resetStruct)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error with provided data: %v", err), code)
			return
		}
		email := resetStruct.Email
		user, err := ctx.UserStore.GetByEmail(email)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting user: %v", err), http.StatusInternalServerError)
			return
		}

		randomID := make([]byte, 32)
		if _, err := rand.Read(randomID); err != nil {
			http.Error(w, "Error generating random ID: %v", http.StatusInternalServerError)
			return
		}
		resetPass := base64.URLEncoding.EncodeToString(randomID)
		auth := smtp.PlainAuth(
			"",
			"resetpassi344@gmail.com",
			"info344!",
			"smtp.gmail.com",
		)
		err = smtp.SendMail(
			"smtp.gmail.com:587",
			auth,
			"resetpassi344@gmail.com",
			[]string{user.Email},
			[]byte("This is the password: "+resetPass),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error sending reset password: %v", err), http.StatusInternalServerError)
			return
		}

		if err = ctx.SessionStore.SavePass(user.Email, resetPass); err != nil {
			http.Error(w, fmt.Sprintf("Error saving reset password: %v", err), http.StatusInternalServerError)
			return
		}
		respond(w, "Password reset sent", http.StatusOK, ContentTypeText)
	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return

	}
}

//CompleteResetHandler uses the one time reset password and resets a new password
func (ctx *Context) CompleteResetHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		vars := mux.Vars(r)
		email := vars["email"]
		resetPass, err := ctx.SessionStore.GetReset(email)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reset password expired: %v", err), http.StatusBadRequest)
			return
		}
		completeReset := &resetInfo{}
		code, err := decodeReq(w, r, completeReset)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error with provided data: %v", err), code)
			return
		}
		if completeReset.Password != completeReset.PasswordConf {
			http.Error(w, "Passwords don't match", http.StatusBadRequest)
			return
		}
		if resetPass != completeReset.ResetPass {
			http.Error(w, "Reset password is wrong", http.StatusBadRequest)
			return
		}

		user, err := ctx.UserStore.GetByEmail(email)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error getting user: %v", err), http.StatusInternalServerError)
			return
		}

		if err = user.SetPassword(completeReset.Password); err != nil {
			http.Error(w, fmt.Sprintf("Error setting password hash: %v", err), http.StatusInternalServerError)
			return
		}

		_, err = ctx.UserStore.UpdatePassword(user.ID, user.PassHash)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error updating password: %v", err), http.StatusInternalServerError)
			return
		}
		respond(w, "New password updated to account", http.StatusOK, ContentTypeText)
	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return

	}

}

//decodeReq checks the header type and decodes the body from the request and
//populates it to the interface returns http.StatusBadRequest if there is an error
func decodeReq(w http.ResponseWriter, r *http.Request, value interface{}) (int, error) {
	if !strings.HasPrefix(r.Header.Get(HeaderContentType), ContentTypeJSON) {
		return http.StatusUnsupportedMediaType, errors.New("Invalid media type")
	}
	if err := json.NewDecoder(r.Body).Decode(value); err != nil {
		return http.StatusBadRequest, fmt.Errorf("error decoding JSON: %v", err)
	}
	return http.StatusOK, nil
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

func getClientKey(r *http.Request) string {
	var ipaddr string
	if r.Header.Get(HeaderForwardedFor) != "" {
		ipaddr = strings.Split(r.Header.Get(HeaderForwardedFor), ",")[0]
	} else {
		ipaddr = r.RemoteAddr
	}
	return ipaddr
}
