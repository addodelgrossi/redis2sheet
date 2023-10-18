package main

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

func PublishEvent(rdb *redis.Client, event EventData, channel string) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return rdb.Publish(channel, data).Err()
}
