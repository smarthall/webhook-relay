package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/smarthall/webhook-relay/internal/messaging"
)

func main() {
	sub := messaging.NewSubscriber()
	msgs, err := sub.Subscribe()
	if err != nil {
		log.Panicf("Failed to consume messages: %s", err)
	}

	client := &http.Client{}

	for msg := range msgs {
		log.Printf("Received message! %s", msg.Body)

		var reqmsg messaging.RequestMessage
		err = json.Unmarshal(msg.Body, &reqmsg)
		if err != nil {
			log.Panicf("Failed to unmarshal message: %s", err)
		}

		req, err := http.NewRequest(reqmsg.Method, "http://localhost:8000"+reqmsg.Path, nil)
		if err != nil {
			log.Panicf("Failed to create request: %s", err)
		}

		for k, v := range reqmsg.Headers {
			req.Header[k] = v
		}

		client.Do(req)
	}
}
