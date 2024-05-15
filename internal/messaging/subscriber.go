package messaging

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Subscriber struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	q    amqp.Queue
}

func NewSubscriber(amqpUri string, key string) *Subscriber {
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

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Panicf("Failed to declare queue: %s", err)
	}

	err = ch.QueueBind(q.Name, key, "webhooks", false, nil)
	if err != nil {
		log.Panicf("Failed to bind queue: %s", err)
	}

	return &Subscriber{
		conn: conn,
		ch:   ch,
		q:    q,
	}
}

func (s *Subscriber) Subscribe() (<-chan amqp.Delivery, error) {
	msgs, err := s.ch.Consume(
		s.q.Name, // queue
		"",       // consumer
		true,     // auto ack
		false,    // exclusive
		false,    // no local
		false,    // no wait
		nil,      // args
	)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}
