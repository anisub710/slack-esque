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
		newUser := &users.NewUser{}
		code, err := decodeReq(w, r, newUser)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error with provided data: %v", err), code)
			return
		}
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
		_, err = sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, stateStruct, w)
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

		email := credentials.Email
		currFails, err := ctx.SessionStore.Increment(email, 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("error saving failed attempts: %v", err), http.StatusInternalServerError)
			return
		}
		if currFails >= 5 {
			ctx.SessionStore.Increment(email, 1)
			currTimeLeft, _ := ctx.SessionStore.TimeLeft(email)
			w.Header().Add(headerRetryAfter, headerRetryAfter)
			http.Error(w, fmt.Sprintf("Too many failed attempts. Try again in %s minutes", currTimeLeft), http.StatusTooManyRequests)
			return
		}

		if err = findUser.Authenticate(credentials.Password); err != nil {
			if _, err := ctx.SessionStore.Increment(email, 1); err != nil {
				http.Error(w, fmt.Sprintf("error saving failed attempts: %v", err), http.StatusInternalServerError)
			}
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		//add to userslogin
		// login := &users.Login{
		// 	Userid:    findUser.ID,
		// 	LoginTime: time.Now(),
		// 	IPAddr:    getClientKey(r),
		// }

		// _, err = ctx.UserStore.InsertLogin(login)
		// if err != nil {
		// 	http.Error(w, fmt.Sprintf("Error inserting login: %v", err), http.StatusInternalServerError)
		// 	return
		// }
		stateStruct := &SessionState{
			BeginTime: time.Now(),
			User:      findUser,
		}
		_, err = sessions.BeginSession(ctx.SigningKey, ctx.SessionStore, stateStruct, w)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error beginning session: %v", err), http.StatusInternalServerError)
			return
		}

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
		vars := mux.Vars(r)
		segment := vars["id"]
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

		respond(w, "Image successfully uploaded", http.StatusOK, contentTypeText)

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
		// splitEmail := strings.Split(email, "@")
		randomID := make([]byte, 32)
		if _, err := rand.Read(randomID); err != nil {
			http.Error(w, "Error generating random ID: %v", http.StatusInternalServerError)
			return
		}
		resetPass := base64.URLEncoding.EncodeToString(randomID)
		//create method in redisstore to set a password (base 64 encoded) that expires in 5 minutes. remove after used or expired

		auth := smtp.PlainAuth(
			"",
			"resetpassi344@gmail.com",
			"info344!",
			"smtp.gmail.com",
		)
		// Connect to the server, authenticate, set the sender and recipient,
		// and send the email all in one step.
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
		// m := gomail.NewMessage()
		// m.SetHeader("From", "resetpassi344@gmail.com")
		// m.SetHeader("To", email)
		// m.SetHeader("Subject", "Password Reset")
		// m.SetBody(contentTypeText, fmt.Sprintf("Hello %s, Here is your one-time password that expires in 5 minutes: %s", user.FullName(), resetPass))

		// d := gomail.NewDialer("smtp.gmail.com", 587, "subramanyamanirudh3@gmail.com", "info344!")
		// if err := d.DialAndSend(m); err != nil {
		// 	http.Error(w, fmt.Sprintf("Error sending email: %v", err), http.StatusInternalServerError)
		// 	return
		// }
		respond(w, "Password reset sent", http.StatusOK, contentTypeText)
	default:
		http.Error(w, "invalid request", http.StatusMethodNotAllowed)
		return

	}
}

//decodeReq checks the header type and decodes the body from the request and
//populates it to the interface returns http.StatusBadRequest if there is an error
func decodeReq(w http.ResponseWriter, r *http.Request, value interface{}) (int, error) {
	if !strings.HasPrefix(r.Header.Get(headerContentType), contentTypeJSON) {
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
	if r.Header.Get(headerForwardedFor) != "" {
		ipaddr = strings.Split(r.Header.Get(headerForwardedFor), ",")[0]
	} else {
		ipaddr = r.RemoteAddr
	}
	return ipaddr
}
