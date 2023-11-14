package main

import (
	"testing"

	"github.com/go-redis/redis"
)

func TestPublishEvent(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// Mock event data
	event := EventData{
		Asset:     "WIN",
		Position:  2,
		Timestamp: 987654322,
		Group:     "test",
		Text:      "This is a test",
		Mode:      "server",
		Name:      "srv02",
	}

	channel := "events"

	// Publish the event
	err := PublishEvent(rdb, event, channel)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Verify the published data
	sub := rdb.Subscribe(channel)
	defer sub.Close()
}

func TestPublishEventSlave(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// Mock event data
	event := EventData{
		Asset:     "WIN",
		Position:  4,
		Timestamp: 987654321,
		Group:     "test",
		Text:      "This is a test",
		Mode:      "slave",
		Name:      "srv01",
	}

	channel := "copy"

	// Publish the event
	err := PublishEvent(rdb, event, channel)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Verify the published data
	sub := rdb.Subscribe(channel)
	defer sub.Close()
}

func TestPublishEventResume(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	// Mock event data
	event := EventData{
		Asset:     "WIN",
		Position:  2,
		Timestamp: 987654321,
		Group:     "test",
		Text:      "This is a test",
		Mode:      "slave",
		Name:      "srv01",
	}

	channel := "resume"

	// Publish the event
	err := PublishEvent(rdb, event, channel)
	if err != nil {
		t.Fatalf("Failed to publish event: %v", err)
	}

	// Verify the published data
	sub := rdb.Subscribe(channel)
	defer sub.Close()
}
