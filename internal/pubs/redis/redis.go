package redis

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"time"
)

type Redis struct {
	connection *redis.Client
	password   string
	address    string
	database   int
	channel    string
}

func (r *Redis) checkConnection() error {
	ctx := context.Background()
	_, err := r.connection.Ping(ctx).Result()
	return err
}

func (r *Redis) connect() {
	r.connection = redis.NewClient(&redis.Options{
		Addr:     r.address,
		Password: r.password,
		DB:       r.database,
	})
}

func (r *Redis) Publish(data interface{}) error {
	repeats := 0
	jsonData, err := json.Marshal(data)
	if err != nil {
		logrus.Errorf("Error marshalling to JSON: %s", err)
		return err
	}
	for {
		ctx := context.Background()
		err = r.connection.Publish(ctx, r.channel, jsonData).Err()
		if err != nil {
			if err := r.checkConnection(); err != nil {
				r.connect()
			}
		} else {
			return nil
		}
		time.Sleep(time.Duration(repeats) * time.Second)
		repeats++
		if repeats > 5 {
			return err
		}
	}
}

func NewRedis(address, password string, db int, channel string) *Redis {
	rdb := &Redis{
		connection: nil,
		password:   password,
		address:    address,
		database:   db,
		channel:    channel,
	}
	return rdb
}

func (r *Redis) TryConnect() {
	for {
		r.connect()
		if err := r.checkConnection(); err != nil {
			logrus.Infof("Connection to redis failed: %v", err)
			time.Sleep(5 * time.Second)
		} else {
			logrus.Infof("Success to connect to redis server")
			return
		}
	}
}
