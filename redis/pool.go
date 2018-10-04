package redis

import (
	"github.com/garyburd/redigo/redis"
	"os"
)

var (
	redisAddress   = ":6379"
	maxConnections = 10
)

func init() {
	if value, ok := os.LookupEnv("REDIS_URL"); ok {
		redisAddress = value
	}
}

// create a new redis pool
func NewPool() *redis.Pool {
	return redis.NewPool(redisConnect, maxConnections)
}

// create new redis connection
func redisConnect() (redis.Conn, error) {
	c, err := redis.Dial("tcp", redisAddress)
	if err != nil {
		panic(err)
	}

	return c, err
}
