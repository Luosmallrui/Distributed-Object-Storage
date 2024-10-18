package redis

import (
	"context"
	"distributed-object-storage/pkg/log"
	"github.com/go-redis/redis/v8"
)

var (
	rdb *redis.Client
)

type Config struct {
	Addr     string `yaml:"addr,omitempty"`
	Password string `yaml:"password,omitempty"`
}

func Init(config *Config) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:       config.Addr,
		Password:   config.Password, // no password set
		MaxRetries: -1,              // Not Retry
		DB:         0,               // use default DB
	})

	pong, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return err
	}
	log.Info(pong, "init redis success.")
	return nil
}

func Redis() *redis.Client {
	if rdb == nil {
		panic("redis not init")
	}
	return rdb
}
