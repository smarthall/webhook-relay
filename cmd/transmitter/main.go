package main

import (
	"log"

	"github.com/smarthall/webhook-relay/internal/messaging"
)

func main() {
	sub := messaging.NewSubscriber()
	msgs, err := sub.Subscribe()
	if err != nil {
		log.Panicf("Failed to consume messages: %s", err)
	}

	for range msgs {
		log.Printf("Received message!")
	}
}
