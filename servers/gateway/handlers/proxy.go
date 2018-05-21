package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"

	"github.com/info344-s18/challenges-ask710/servers/gateway/sessions"
)

//NewServiceProxy returns a new ReverseProxy
//for a microservice given a comma-delimited
//list of network addresses
func (ctx *Context) NewServiceProxy(addrs string) *httputil.ReverseProxy {
	splitAddrs := strings.Split(addrs, ",")
	nextAddr := 0
	mx := sync.Mutex{}

	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.URL.Scheme = "http"
			mx.Lock()
			r.URL.Host = splitAddrs[nextAddr]
			nextAddr = (nextAddr + 1) % len(splitAddrs)
			mx.Unlock()

			r.Header.Del(HeaderUser)
			stateStruct := &SessionState{}
			_, err := sessions.GetState(r, ctx.SigningKey, ctx.SessionStore, stateStruct)
			if err != nil {
				return
			}
			userJSON, err := json.Marshal(stateStruct.User)
			if err != nil {
				return
			}
			r.Header.Set(HeaderUser, string(userJSON))

		},
	}
}
