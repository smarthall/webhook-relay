package main

import (
	"log"
	"net/http"
	"time"

	"github.com/smarthall/webhook-relay/internal/publisher"
)

var pub = publisher.New()

func myHandler(w http.ResponseWriter, r *http.Request) {
	err := pub.Publish(*r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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
