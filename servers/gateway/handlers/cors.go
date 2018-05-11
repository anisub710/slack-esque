package handlers

import (
	"fmt"
	"net/http"
)

/* TODO: implement a CORS middleware handler, as described
in https://drstearns.github.io/tutorials/cors/ that responds
with the following headers to all requests:

  Access-Control-Allow-Origin: *
  Access-Control-Allow-Methods: GET, PUT, POST, PATCH, DELETE
  Access-Control-Allow-Headers: Content-Type, Authorization
  Access-Control-Expose-Headers: Authorization
  Access-Control-Max-Age: 600
*/

//CorsHandler is a middleware handler
type CorsHandler struct {
	handler http.Handler
}

//NewCorsHandler constructs a new CorsHandler middleware handler
func NewCorsHandler(handler http.Handler) *CorsHandler {
	return &CorsHandler{handler}
}

//HandleCors adds necessary headers to the handler
func (c *CorsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	methods := fmt.Sprintf("%s, %s, %s, %s, %s", http.MethodGet, http.MethodPut,
		http.MethodPost, http.MethodPatch, http.MethodDelete)

	w.Header().Add(HeaderAccessControlAllowOrigin, "*")
	w.Header().Add(HeaderAccessControlAllowMethods, methods)
	w.Header().Add(HeaderAccessControlAllowHeaders, AllowHeadersAuth)
	w.Header().Add(HeaderAccessControlExposeHeaders, ExposeHeadersAuth)
	w.Header().Add(HeaderAccessControlMaxAge, MaxAge)
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusOK)
	default:
		c.handler.ServeHTTP(w, r)
	}

}
