package main

import (
	"log"
	"net/http"
	"time"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("Received request")
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
