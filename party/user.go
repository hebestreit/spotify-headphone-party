package party

import (
	"encoding/json"
	"fmt"
	"github.com/hebestreit/spotify-headphone-party/redis"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"
)

type User struct {
	ID    string
	Token *oauth2.Token
}

func (u *User) SpotifyClient() *spotify.Client {
	client := auth.NewClient(u.Token)
	return &client
}

// create new user object and persist in redis
func CreateUser(userID string, token *oauth2.Token) *User {
	user := &User{ID: userID, Token: token}
	data, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}

	pool := redis.NewPool()
	defer pool.Close()

	c := pool.Get()

	_, err = c.Do("SET", "user:"+user.ID, data)
	if err != nil {
		panic(err)
	}

	return user
}

// retrieve user object from redis
func FindUser(userId string) (*User, error) {
	pool := redis.NewPool()
	defer pool.Close()

	c := pool.Get()

	var user User
	reply, err := c.Do("GET", fmt.Sprintf("user:%s", userId))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(reply.([]byte), &user); err != nil {
		return nil, err
	}

	return &user, nil
}
