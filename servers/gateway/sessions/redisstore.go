package sessions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

//RedisStore represents a session.Store backed by redis.
type RedisStore struct {
	//Redis client used to talk to redis server.
	Client *redis.Client
	//Used for key expiry time on redis.
	SessionDuration time.Duration
}

//NewRedisStore constructs a new RedisStore
func NewRedisStore(client *redis.Client, sessionDuration time.Duration) *RedisStore {
	//initialize and return a new RedisStore struct

	//Test for null client and time duration

	return &RedisStore{
		Client:          client,
		SessionDuration: sessionDuration,
	}
}

//Store implementation

//Save saves the provided `sessionState` and associated SessionID to the store.
//The `sessionState` parameter is typically a pointer to a struct containing
//all the data you want to associated with the given SessionID.
func (rs *RedisStore) Save(sid SessionID, sessionState interface{}) error {
	//TODO: marshal the `sessionState` to JSON and save it in the redis database,
	//using `sid.getRedisKey()` for the key.
	//return any errors that occur along the way.
	j, err := json.Marshal(sessionState)
	if err != nil {
		return fmt.Errorf("Error marshaling session state: %v", err)
	}

	err = rs.Client.Set(sid.getRedisKey(), j, rs.SessionDuration).Err()
	if err != nil {
		return fmt.Errorf("Error saving session data in redis: %v", err)
	}
	return nil
}

//Get populates `sessionState` with the data previously saved
//for the given SessionID
func (rs *RedisStore) Get(sid SessionID, sessionState interface{}) error {
	//TODO: get the previously-saved session state data from redis,
	//unmarshal it back into the `sessionState` parameter
	//and reset the expiry time, so that it doesn't get deleted until
	//the SessionDuration has elapsed.

	//for extra-credit using the Pipeline feature of the redis
	//package to do both the get and the reset of the expiry time
	//in just one network round trip!

	pipeline := rs.Client.Pipeline()
	getPipe := pipeline.Get(sid.getRedisKey())
	pipeline.Expire(sid.getRedisKey(), rs.SessionDuration)
	_, err := pipeline.Exec()
	if err != nil {
		return ErrStateNotFound
	}

	prevState, err := getPipe.Result()
	if err != nil {
		return ErrStateNotFound
	}

	err = json.Unmarshal([]byte(prevState), sessionState)
	if err != nil {
		return fmt.Errorf("Error unmarshaling session state: %v", err)
	}

	return nil
}

//Delete deletes all state data associated with the SessionID from the store.
func (rs *RedisStore) Delete(sid SessionID) error {
	//TODO: delete the data stored in redis for the provided SessionID

	if err := rs.Client.Del(sid.getRedisKey()).Err(); err != nil {
		return fmt.Errorf("Error deleting: %v", err)
	}

	return nil
}

//Increment increments the number of failed attempts to sign in
func (rs *RedisStore) Increment(id string, by int64) (int64, error) {
	currFails, err := rs.Client.IncrBy(id, by).Result()
	if err != nil {
		return 0, err
	}
	if currFails == 5 {
		if _, err := rs.Client.Expire(id, 10*time.Minute).Result(); err != nil {
			return 0, err
		}
	}
	return currFails, nil
}

//TimeLeft returns the time left for the block to be lifted
func (rs *RedisStore) TimeLeft(id string) (string, error) {
	currTimeLeft, err := rs.Client.TTL(id).Result()
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(currTimeLeft.Minutes(), 'f', 1, 64), nil
}

//SavePass saves the reset password for an email
func (rs *RedisStore) SavePass(email string, resetPass string) error {
	if _, err := rs.Client.Set(email, resetPass, 5*time.Minute).Result(); err != nil {
		return err
	}
	return nil
}

//GetReset gets the reset password for an email
func (rs *RedisStore) GetReset(email string) (string, error) {
	return rs.Client.Get(email).Result()
}

//getRedisKey() returns the redis key to use for the SessionID
func (sid SessionID) getRedisKey() string {
	//convert the SessionID to a string and add the prefix "sid:" to keep
	//SessionID keys separate from other keys that might end up in this
	//redis instance
	return "sid:" + sid.String()
}
