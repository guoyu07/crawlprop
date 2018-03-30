package stored

import (
	"github.com/go-redis/redis"
	"github.com/millken/crawlprop/config"
)

var client *redis.Client

func Initialize(red config.RedisConfig) error {
	var err error
	client = redis.NewClient(&redis.Options{
		Addr:     red.Addr,
		Password: "",
		DB:       0,
	})

	_, err = client.Ping().Result()

	return err
}

func RedisClient() *redis.Client {
	return client
}
