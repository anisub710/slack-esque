package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

//middleware for blocking subsequent sign-ins

//Blocker is a struct for blocking handler
type Blocker struct {
	wrappedHandler http.Handler
	redisClient    *redis.Client
	maxFails       int64
	rejectTime     time.Duration
}

//TrackLogin wraps handlerToWrap with a Blocker handler.
func TrackLogin(handlerToWrap http.Handler, redisClient *redis.Client, maxFails int64,
	rejectTime time.Duration) http.Handler {

	return &Blocker{
		wrappedHandler: handlerToWrap,
		redisClient:    redisClient,
		maxFails:       maxFails,
		rejectTime:     rejectTime,
	}
}

func getClientKey(r *http.Request) string {
	var ipaddr string
	if r.Header.Get(headerForwardedFor) != "" {
		ipaddr = strings.Split(r.Header.Get(headerForwardedFor), ",")[0]
	} else {
		ipaddr = r.RemoteAddr
	}
	return ipaddr
}

func (b *Blocker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// ipaddr := getClientKey(r)
	// currFails, _ := b.redisClient.Incr(ipaddr).Result()
}
