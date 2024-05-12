package messaging

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(amqpUri string) *Publisher {
	var err error

	conn, err := amqp.Dial(amqpUri)
	if err != nil {
		log.Panicf("Failed to connect to RabbitMQ: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Panicf("Failed to open a channel: %s", err)
	}

	err = ch.ExchangeDeclare(
		"webhooks", // name
		"topic",    // type
		true,       // durable
		false,      // auto-deleted
		false,      // internal
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Panicf("Failed to declare exchange: %s", err)
	}

	return &Publisher{
		conn: conn,
		ch:   ch,
	}
}

func (p *Publisher) Publish(request http.Request) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return err
	}

	msg := RequestMessage{
		Method:  request.Method,
		Path:    request.URL.Path,
		Headers: request.Header,
		Body:    string(body),
	}

	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = p.ch.PublishWithContext(ctx,
		"webhooks", // exchange
		"test",     // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        json,
		})

	return err
}
