package messaging

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewPublisher(amqpUri string) *Publisher {
	var err error
	// ensure pooled connections are initialized
	if err = InitConnections(amqpUri); err != nil {
		log.Panicf("Failed to initialize RabbitMQ connections: %s", err)
	}

	conn := GetPubConn()
	if conn == nil {
		log.Panicf("publisher connection is nil")
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

func (p *Publisher) Publish(msg RequestMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	json, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	routingKey := strings.Trim(strings.Replace(msg.Path, "/", ".", -1), ".")
	err = p.ch.PublishWithContext(ctx,
		"webhooks", // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        json,
		})
	log.Printf("Published message to %s", routingKey)

	return err
}
