package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/info344-s18/challenges-ask710/servers/gateway/indexes"
	"github.com/info344-s18/challenges-ask710/servers/gateway/models/users"
	"github.com/streadway/amqp"

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
	messageAddrs := reqEnv("MESSAGESADDR")
	summaryAddrs := reqEnv("SUMMARYADDR")
	mqAddr := reqEnv("MQADDR")
	mqName := reqEnv("MQNAME")
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
		log.Printf("Error connecting to redis database: %v", err)
		os.Exit(1)
	}

	redisStore := sessions.NewRedisStore(redisClient, time.Hour)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	defer db.Close()
	userStore := users.NewMySQLStore(db)

	trie, err := userStore.LoadUsers()
	if err != nil {
		trie = indexes.NewTrie()
	}

	conn, err := connectToMQ(mqAddr)
	if err != nil {
		log.Printf("error dialing MQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Printf("error getting channel: %v", err)
	}

	q, err := channel.QueueDeclare(mqName,
		false,
		false,
		false,
		false,
		nil)
	if err != nil {
		log.Printf("error declaring queue: %v", err)
	}

	messages, err := channel.Consume(q.Name,
		"",
		false,
		false,
		false,
		false,
		nil)

	if err != nil {
		log.Printf("error consuming messages: %v", err)
	}
	notifier := handlers.NewNotifier()
	ctx := handlers.NewContext(sessionKey, redisStore, userStore, trie, notifier)

	go ctx.Notifier.ProcessMessages(messages)

	mux := mux.NewRouter()

	mux.HandleFunc("/v1/users", ctx.UsersHandler)
	mux.HandleFunc("/v1/users/{id}", ctx.SpecificUserHandler)
	mux.HandleFunc("/v1/sessions", ctx.SessionsHandler)
	mux.HandleFunc("/v1/sessions/{id}", ctx.SpecificSessionHandler)
	mux.HandleFunc("/v1/users/{id}/avatar", ctx.AvatarHandler)
	mux.HandleFunc("/v1/resetcodes", ctx.ResetHandler)
	mux.HandleFunc("/v1/passwords/{email}", ctx.CompleteResetHandler)

	mux.Handle("/v1/summary", ctx.NewServiceProxy(summaryAddrs))

	messageService := ctx.NewServiceProxy(messageAddrs)
	mux.Handle("/v1/channels", messageService)
	mux.Handle("/v1/channels/{channelID}", messageService)
	mux.Handle("/v1/channels/{channelID}/members", messageService)
	mux.Handle("/v1/messages/{messageID}", messageService)

	mux.Handle("/v1/ws", handlers.NewWebSocketHandler(ctx))
	wrappedMux := handlers.NewCorsHandler(mux)

	log.Printf("Server is listening at https://%s", addr)
	log.Fatal(http.ListenAndServeTLS(addr, tlsCertPath, tlsKeyPath, wrappedMux))
}

func reqEnv(name string) string {
	val := os.Getenv(name)
	if len(val) == 0 {
		//Fatal?
		log.Fatalf("Please set %s variable", name)
		os.Exit(1)
	}
	return val
}

//connectToMQ makes retries if necessary to connect to rabbitmq
func connectToMQ(addr string) (*amqp.Connection, error) {
	mqURL := "amqp://" + addr
	var conn *amqp.Connection
	var err error
	for i := 1; i <= handlers.MaxConnRetries; i++ {
		conn, err = amqp.Dial(mqURL)
		if err == nil {
			log.Printf("successfully connect to %s", mqURL)
			return conn, nil
		}
		log.Printf("error connecting to MQ at %s: %v", mqURL, err)
		log.Printf("will retry in %d seconds", i*2)
		time.Sleep(time.Second * time.Duration(i*2))
	}
	return nil, err
}
