package handlers

import (
	"bytes"
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"

	"github.com/go-sql-driver/mysql"
	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"
)

func TestUsersHandler(t *testing.T) {
	cases := []struct {
		name                string
		jsonReq             string
		expectedStatusCode  int
		expectedContentType string
		expectedHeaderType  string
	}{
		{
			"Valid new user",
			`{"email": "test1@uw.edu","password":"test1234","passwordConf":"test1234","userName":"test1","firstName":"Competent","lastName": "Gopher"}`,
			http.StatusCreated,
			contentTypeJSON,
			contentTypeJSON,
		},
	}

	for _, c := range cases {
		URL := "/v1/users"
		json := []byte(c.jsonReq)
		req, err := http.NewRequest("POST", URL, bytes.NewBuffer(json))
		req.Header.Set("Content-Type", "application/json")
		respRec := httptest.NewRecorder()

		sessionStore := sessions.NewMemStore(time.Hour, time.Minute)
		// db, _, err := sqlmock.New()
		// userStore := users.NewMySQLStore(db)
		// if err != nil {
		// 	t.Fatalf("error creating sql mock: %v", err)
		// }

		// defer db.Close()
		config := mysql.Config{
			Addr:   "0.0.0.0:3306",
			User:   "root",
			Passwd: "password",
			DBName: "users",
		}
		db, err := sql.Open("mysql", config.FormatDSN())
		if err != nil {
			log.Fatalf("error opening database: %v", err)
		}
		userStore := users.NewMySQLStore(db)
		ctx := NewContext("test key", sessionStore, userStore)
		ctx.UsersHandler(respRec, req)
		t.Errorf(respRec.Body.String())
		resp := respRec.Result()
		if resp.StatusCode != c.expectedStatusCode {
			t.Errorf("case %s: incorrect status code: expected %d but got %d",
				c.name, c.expectedStatusCode, resp.StatusCode)
		}
		allowedOrigin := resp.Header.Get(headerAccessControlAllowOrigin)
		if allowedOrigin != "*" {
			t.Errorf("case %s: incorrect CORS header: expected %s but got %s",
				c.name, "*", allowedOrigin)
		}

		contentType := resp.Header.Get(headerContentType)
		if contentType != c.expectedContentType {
			t.Errorf("case %s: incorrect Content-Type header: expected %s but got %s",
				c.name, c.expectedContentType, contentType)
		}

	}
}
