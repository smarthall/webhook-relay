package main

import (
	"log"
	"net/http"
	"time"

	"github.com/smarthall/webhook-relay/internal/messaging"
)

var pub = messaging.NewPublisher()

func myHandler(w http.ResponseWriter, r *http.Request) {
	// Publish the request to the message broker
	err := pub.Publish(*r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// If the message was published successfully, return a 204 No Content response
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	s := &http.Server{
		Addr:           ":8080",
		Handler:        http.HandlerFunc(myHandler),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
