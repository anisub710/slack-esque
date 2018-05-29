package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	addr := os.Getenv("ADDR")
	if len(addr) == 0 {
		addr = ":80"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/summary", SummaryHandler)

	log.Fatal(http.ListenAndServe(addr, mux))

}
