package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCorsHandler(t *testing.T) {
	newHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	server := httptest.NewServer(NewCorsHandler(newHandler))
	var u bytes.Buffer
	u.WriteString(string(server.URL))
	u.WriteString("/v1/test")
	res, _ := http.Get(u.String())

	if res.Header.Get(HeaderAccessControlAllowOrigin) != "*" {
		t.Errorf("Access-Control-Allow-Origin header not set")
	}

	methods := fmt.Sprintf("%s, %s, %s, %s, %s", http.MethodGet, http.MethodPut,
		http.MethodPost, http.MethodPatch, http.MethodDelete)
	if res.Header.Get(HeaderAccessControlAllowMethods) != methods {
		t.Errorf("Access-Control-Allow-Methods header not set")
	}

	if res.Header.Get(HeaderAccessControlAllowHeaders) != AllowHeadersAuth {
		t.Errorf("Access-Control-Allow-Headers header not set")
	}
	if res.Header.Get(HeaderAccessControlExposeHeaders) != ExposeHeadersAuth {
		t.Errorf("Access-Control-Expose-Headers header not set")
	}
	if res.Header.Get(HeaderAccessControlMaxAge) != MaxAge {
		t.Errorf("Access-Control-Max-Age header not set")
	}
}
