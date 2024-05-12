package publisher

import (
	"context"
	"log"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func New() *Publisher {
	var err error

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	err := p.ch.PublishWithContext(ctx,
		"webhooks", // exchange
		"test",     // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte("Hello World!"),
		})

	return err
}
