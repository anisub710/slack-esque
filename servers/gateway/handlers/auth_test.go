package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"

	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"
)

const userURL = "/v1/users"
const specUserURL = "/v1/users/"
const sessionURL = "/v1/sessions"

type VarsHandler func(w http.ResponseWriter, r *http.Request, vars map[string]string)

func (vh VarsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	vh(w, r, vars)
}

func createTestUser(userType string) *users.User {
	var expectedUser *users.User
	switch userType {
	case "normal":
		expectedUser = &users.User{
			ID:        1,
			UserName:  "test1",
			FirstName: "Competent",
			LastName:  "Gopher",
			PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
		}

	case "new":
		expectedUser = &users.User{
			ID:        1,
			Email:     "test1@uw.edu",
			UserName:  "test1",
			FirstName: "Competent",
			LastName:  "Gopher",
			PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
		}
		expectedUser.SetPassword("test1234")
	case "insertError":
		expectedUser = &users.User{
			ID:        1,
			Email:     "test123@uw.edu",
			PassHash:  nil,
			UserName:  "competentGopher",
			FirstName: "Competent",
			LastName:  "Gopher",
			PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
		}
	case "updated":
		expectedUser = &users.User{
			ID:        1,
			Email:     "test123@uw.edu",
			PassHash:  nil,
			UserName:  "competentGopher",
			FirstName: "Incompetent",
			LastName:  "Shark",
			PhotoURL:  "https://www.gravatar.com/avatar/9ed8dc990d56d07d330e5a057254cca9",
		}

	}

	return expectedUser
}

func TestUsersHandler(t *testing.T) {
	cases := []struct {
		name                string
		jsonReq             string
		expectedStatusCode  int
		expectedContentType string
		userStore           *users.MockStore
		method              string
		setContentType      string
		signingKey          string
	}{
		{
			"Valid new user",
			`{"email": "test1@uw.edu",
				"password":"test1234",
				"passwordConf":"test1234",
				"userName":"test1",
				"firstName":"Competent",
				"lastName": "Gopher"}`,
			http.StatusCreated,
			contentTypeJSON,
			&users.MockStore{
				TriggerError: false,
				Result:       createTestUser("normal"),
			},
			http.MethodPost,
			contentTypeJSON,
			"test key",
		},
		{
			"Invalid content type header",
			`{"email": "test1@uw.edu",
				"password":"test1234",
				"passwordConf":"test1234",
				"userName":"test1",
				"firstName":"Competent",
				"lastName": "Gopher"}`,
			http.StatusUnsupportedMediaType,
			contentTypeText,
			&users.MockStore{},
			http.MethodPost,
			contentTypeHTML,
			"test key",
		},
		{
			"Invalid json to decode",
			`{,
				"password":"test1234",
				"passwordConf":"test1234",
				"userName":"test1",
				"firstName":"Competent",
				"lastName": "Gopher"}`,
			http.StatusBadRequest,
			contentTypeText,
			&users.MockStore{},
			http.MethodPost,
			contentTypeJSON,
			"test key",
		},
		{
			"Invalid User",
			`{"email": "test1wdu",
				"password":"test1234",
				"passwordConf":"test1234",
				"userName":"test1",
				"firstName":"Competent",
				"lastName": "Gopher"}`,
			http.StatusInternalServerError,
			contentTypeText,
			&users.MockStore{},
			http.MethodPost,
			contentTypeJSON,
			"test key",
		},
		{
			"Invalid Insert",
			`{"email": "test1@uw.edu",
				"password":"test1234",
				"passwordConf":"test1234",
				"userName":"test1",
				"firstName":"Competent",
				"lastName": "Gopher"}`,
			http.StatusInternalServerError,
			contentTypeText,
			&users.MockStore{
				TriggerError: true,
			},
			http.MethodPost,
			contentTypeJSON,
			"test key",
		},
		{
			"Invalid Begin Session",
			`{"email": "test1@uw.edu",
				"password":"test1234",
				"passwordConf":"test1234",
				"userName":"test1",
				"firstName":"Competent",
				"lastName": "Gopher"}`,
			http.StatusInternalServerError,
			contentTypeText,
			&users.MockStore{},
			http.MethodPost,
			contentTypeJSON,
			"",
		},
		{
			"Invalid Method",
			`{"email": "test1@uw.edu",
				"password":"test1234",
				"passwordConf":"test1234",
				"userName":"test1",
				"firstName":"Competent",
				"lastName": "Gopher"}`,
			http.StatusMethodNotAllowed,
			contentTypeText,
			&users.MockStore{},
			http.MethodGet,
			contentTypeJSON,
			"test key",
		},
	}

	for _, c := range cases {
		bytesJSON := []byte(c.jsonReq)
		queryJSON := bytes.NewBuffer(bytesJSON)
		req, err := http.NewRequest(c.method, userURL, queryJSON)
		if err != nil {
			t.Errorf("case %s: Error sending request: %v", c.name, err)
		}
		req.Header.Set("Content-Type", c.setContentType)
		respRec := httptest.NewRecorder()

		sessionStore := sessions.NewMemStore(time.Hour, time.Minute)

		ctx := NewContext(c.signingKey, sessionStore, c.userStore)
		ctx.UsersHandler(respRec, req)
		// t.Errorf(respRec.Body.String())
		resp := respRec.Result()
		if resp.StatusCode != c.expectedStatusCode {
			t.Errorf("case %s: incorrect status code: expected %d but got %d",
				c.name, c.expectedStatusCode, resp.StatusCode)
		}
		resultUser := &users.User{}
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			if err = json.Unmarshal(respRec.Body.Bytes(), resultUser); err != nil {
				t.Errorf("case %s: Error unmarshalling json: %v", c.name, err)
			}

			if !reflect.DeepEqual(c.userStore.Result, resultUser) {
				t.Errorf("case %s: Result not equal to expected result", c.name)
			}
		}

		// allowedOrigin := resp.Header.Get(headerAccessControlAllowOrigin)
		// if allowedOrigin != "*" {
		// 	t.Errorf("case %s: incorrect CORS header: expected %s but got %s",
		// 		c.name, "*", allowedOrigin)
		// }

		contentType := resp.Header.Get(headerContentType)
		if !strings.Contains(contentType, c.expectedContentType) {
			t.Errorf("case %s: incorrect Content-Type header: expected %s but got %s",
				c.name, c.expectedContentType, contentType)
		}

	}
}

func getSessionID(signingKey string) sessions.SessionID {
	id, _ := sessions.NewSessionID(signingKey)
	return id
}

// func TestSpecificUserHandler(t *testing.T) {
// 	cases := []struct {
// 		name                string
// 		jsonReq             string
// 		expectedStatusCode  int
// 		expectedContentType string
// 		userStore           *users.MockStore
// 		method              string
// 		setContentType      string
// 		signingKey          string
// 		id                  string
// 		sesssionID          sessions.SessionID
// 	}{
// 		{
// 			"Get Valid user me",
// 			"",
// 			http.StatusOK,
// 			contentTypeJSON,
// 			// contentTypeJSON,
// 			&users.MockStore{
// 				TriggerError: false,
// 				Result:       createTestUser("normal"),
// 			},
// 			http.MethodGet,
// 			contentTypeJSON,
// 			"test key",
// 			"{me}",
// 			getSessionID("test key"),
// 		},
// 	}

// 	for _, c := range cases {
// 		URL := specUserURL + c.id
// 		// bytesJSON := []byte{}
// 		// if c.method == http.MethodPatch {
// 		// 	bytesJSON = []byte(c.jsonReq)
// 		// }
// 		// queryJSON := bytes.NewBuffer(bytesJSON)
// 		req, err := http.NewRequest(c.method, URL, nil)
// 		if err != nil {
// 			t.Errorf("case %s: Error sending request: %v", c.name, err)
// 		}
// 		// req.Header.Set("Content-Type", c.setContentType)
// 		req.Header.Set("Authorization", "Bearer "+c.sesssionID.String())
// 		respRec := httptest.NewRecorder()

// 		sessionStore := sessions.NewMemStore(time.Hour, time.Minute)
// 		sessionStore.Save(c.sesssionID, c.userStore.Result)
// 		ctx := NewContext(c.signingKey, sessionStore, c.userStore)

// 		mainRouter := mux.NewRouter().StrictSlash(true)
// 		mainRouter.HandleFunc("/v1/users/{id}", VarsHandler(ctx.SpecificUserHandler)).Name(URL).Methods("GET")
// 		// t.Errorf(respRec.Body.String())
// 		resp := respRec.Result()
// 		if resp.StatusCode != c.expectedStatusCode {
// 			t.Errorf("case %s: incorrect status code: expected %d but got %d: %s",
// 				c.name, c.expectedStatusCode, resp.StatusCode, respRec.Body.String())
// 		}
// 		resultUser := &users.User{}
// 		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
// 			if err = json.Unmarshal(respRec.Body.Bytes(), resultUser); err != nil {
// 				t.Errorf("case %s: Error unmarshalling json: %v", c.name, err)
// 			}

// 			if !reflect.DeepEqual(c.userStore.Result, resultUser) {
// 				t.Errorf("case %s: Result not equal to expected result", c.name)
// 			}
// 		}

// 		// allowedOrigin := resp.Header.Get(headerAccessControlAllowOrigin)
// 		// if allowedOrigin != "*" {
// 		// 	t.Errorf("case %s: incorrect CORS header: expected %s but got %s",
// 		// 		c.name, "*", allowedOrigin)
// 		// }

// 		contentType := resp.Header.Get(headerContentType)
// 		if !strings.Contains(contentType, c.expectedContentType) {
// 			t.Errorf("case %s: incorrect Content-Type header: expected %s but got %s",
// 				c.name, c.expectedContentType, contentType)
// 		}

// 	}
// }

func TestSessionsHandler(t *testing.T) {
	cases := []struct {
		name                string
		jsonReq             string
		expectedStatusCode  int
		expectedContentType string
		userStore           *users.MockStore
		method              string
		setContentType      string
		signingKey          string
	}{
		{
			"Valid Credentials",
			`{
				"email": "test1@uw.edu",
				"password":"test1234"
			}`,
			http.StatusCreated,
			contentTypeJSON,
			&users.MockStore{
				TriggerError: false,
				Result:       createTestUser("new"),
			},
			http.MethodPost,
			contentTypeJSON,
			"test key",
		},
		{
			"Invalid Content Type Header",
			`{
				"email": "test1@uw.edu",
				"password":"test1234"
			}`,
			http.StatusUnsupportedMediaType,
			contentTypeText,
			&users.MockStore{
				TriggerError: false,
				Result:       createTestUser("new"),
			},
			http.MethodPost,
			contentTypeHTML,
			"test key",
		},
		{
			"Invalid JSON ",
			`{		,		
				"password":"test1234"
			}`,
			http.StatusBadRequest,
			contentTypeText,
			&users.MockStore{
				TriggerError: false,
				Result:       createTestUser("new"),
			},
			http.MethodPost,
			contentTypeJSON,
			"test key",
		},
		{
			"Invalid GetByEmail ",
			`{	
				"email": "test1@uw.edu",
				"password":"test1234"
			}`,
			http.StatusUnauthorized,
			contentTypeText,
			&users.MockStore{
				TriggerError: true,
			},
			http.MethodPost,
			contentTypeJSON,
			"test key",
		},
		{
			"Invalid Password ",
			`{	
				"email": "test1@uw.edu",
				"password":"test123456"
			}`,
			http.StatusUnauthorized,
			contentTypeText,
			&users.MockStore{
				TriggerError: false,
				Result:       createTestUser("new"),
			},
			http.MethodPost,
			contentTypeJSON,
			"test key",
		},
		{
			"Invalid Begin Session",
			`{
				"email": "test1@uw.edu",
				"password":"test1234"
			}`,
			http.StatusInternalServerError,
			contentTypeText,
			&users.MockStore{
				TriggerError: false,
				Result:       createTestUser("new"),
			},
			http.MethodPost,
			contentTypeJSON,
			"",
		},
		{
			"Invalid Method",
			`{
				"email": "test1@uw.edu",
				"password":"test1234"
			}`,
			http.StatusMethodNotAllowed,
			contentTypeText,
			&users.MockStore{
				TriggerError: false,
				Result:       createTestUser("new"),
			},
			http.MethodGet,
			contentTypeJSON,
			"test key",
		},
	}

	for _, c := range cases {
		bytesJSON := []byte(c.jsonReq)
		queryJSON := bytes.NewBuffer(bytesJSON)
		req, err := http.NewRequest(c.method, userURL, queryJSON)
		if err != nil {
			t.Errorf("case %s: Error sending request: %v", c.name, err)
		}
		req.Header.Set("Content-Type", c.setContentType)
		respRec := httptest.NewRecorder()

		sessionStore := sessions.NewMemStore(time.Hour, time.Minute)

		ctx := NewContext(c.signingKey, sessionStore, c.userStore)
		ctx.SessionsHandler(respRec, req)
		// t.Errorf(respRec.Body.String())
		resp := respRec.Result()
		if resp.StatusCode != c.expectedStatusCode {
			t.Errorf("case %s: incorrect status code: expected %d but got %d: %s",
				c.name, c.expectedStatusCode, resp.StatusCode, respRec.Body.String())
		}
		resultUser := &users.User{}
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			if err = json.Unmarshal(respRec.Body.Bytes(), resultUser); err != nil {
				t.Errorf("case %s: Error unmarshalling json: %v", c.name, err)
			}
			if !reflect.DeepEqual(createTestUser("normal"), resultUser) {
				t.Errorf("case %s: Result not equal to expected result", c.name)
			}
		}

		// allowedOrigin := resp.Header.Get(headerAccessControlAllowOrigin)
		// if allowedOrigin != "*" {
		// 	t.Errorf("case %s: incorrect CORS header: expected %s but got %s",
		// 		c.name, "*", allowedOrigin)
		// }

		contentType := resp.Header.Get(headerContentType)
		if !strings.Contains(contentType, c.expectedContentType) {
			t.Errorf("case %s: incorrect Content-Type header: expected %s but got %s",
				c.name, c.expectedContentType, contentType)
		}

	}
}
