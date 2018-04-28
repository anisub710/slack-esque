package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func respond(w http.ResponseWriter, value interface{}, statusCode int, contentType string) {
	w.Header().Add(headerContentType, contentType)
	w.WriteHeader(statusCode)

	switch contentType {
	case contentTypeJSON:
		if err := json.NewEncoder(w).Encode(value); err != nil {
			log.Printf("Error encoding JSON: %v", err)
		}
	case contentTypeText:
		w.Write([]byte(value.(string)))
	}

}
