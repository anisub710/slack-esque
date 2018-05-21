package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func respond(w http.ResponseWriter, value interface{}, statusCode int, contentType string) {
	w.Header().Add(HeaderContentType, contentType)
	w.WriteHeader(statusCode)

	switch contentType {
	case ContentTypeJSON:
		if err := json.NewEncoder(w).Encode(value); err != nil {
			log.Printf("Error encoding JSON: %v", err)
		}
	case ContentTypeText:
		w.Write([]byte(value.(string)))
	}

}
