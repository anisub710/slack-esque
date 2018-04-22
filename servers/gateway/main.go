package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"

	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/info344-s18/challenges-ask710/servers/gateway/handlers"
)

//main is the main entry point for the server
func main() {
	/* TODO: add code to do the following
	- Read the ADDR environment variable to get the address
	  the server should listen on. If empty, default to ":80"
	- Create a new mux for the web server.
	- Tell the mux to call your handlers.SummaryHandler function
	  when the "/v1/summary" URL path is requested.
	- Start a web server listening on the address you read from
	  the environment variable, using the mux you created as
	  the root handler. Use log.Fatal() to report any errors
	  that occur when trying to start the web server.
	*/

	addr := os.Getenv("ADDR")
	sessionKey := reqEnv("SESSIONKEY")
	redisAddr := reqEnv("REDISADDR")
	dsn := reqEnv("DSN")

	if len(addr) == 0 {
		addr = ":443"
	}

	tlsKeyPath := reqEnv("TLSKEY")
	tlsCertPath := reqEnv("TLSCERT")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping().Result()
	if err != nil {
		//do something
	}

	//check time duration
	redisStore := sessions.NewRedisStore(redisClient, time.Hour)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	defer db.Close()
	userStore := users.NewMySQLStore(db)

	ctx := handlers.NewContext(sessionKey, redisStore, userStore)

	mux := mux.NewRouter()

	mux.HandleFunc("/v1/summary", handlers.SummaryHandler)
	mux.HandleFunc("/v1/users", ctx.UsersHandler)
	mux.HandleFunc("/v1/users/", ctx.SpecificUserHandler)
	mux.HandleFunc("/v1/sessions", ctx.SessionsHandler)
	mux.HandleFunc("/v1/sessions/", ctx.SpecificSessionHandler)
	mux.HandleFunc("/v1/users/{id}/avatar", ctx.AvatarHandler)

	wrappedMux := handlers.NewCorsHandler(mux)

	log.Printf("Server is listening at https://%s", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlsCertPath, tlsKeyPath, wrappedMux))
}

func reqEnv(name string) string {
	val := os.Getenv(name)
	if len(val) == 0 {
		log.Fatalf("Please set %s variable", name)
		os.Exit(1)
	}
	return val
}
