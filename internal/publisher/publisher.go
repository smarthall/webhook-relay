package publisher

import "net/http"

type Publisher struct{}

func New() *Publisher {
	return &Publisher{}
}

func (p *Publisher) Publish(request http.Request) error {
	// Publish the message

	return nil
}
