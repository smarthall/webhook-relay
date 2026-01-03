package messaging

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// HealthChecker publishes heartbeat messages to an instance-specific routing key
// and subscribes to a temporary exclusive queue bound to that key. If a published
// heartbeat is not observed within the configured timeout, a notification is sent
// on the Failure channel.
type HealthChecker struct {
	amqpUri  string
	Interval time.Duration
	Timeout  time.Duration

	Failure chan struct{}
	chPub   *amqp.Channel
	chSub   *amqp.Channel

	stop chan struct{}
}

func NewHealthChecker(amqpUri string, interval, timeout time.Duration) *HealthChecker {
	h := &HealthChecker{
		amqpUri:  amqpUri,
		Interval: interval,
		Timeout:  timeout,
		Failure:  make(chan struct{}, 1),
		stop:     make(chan struct{}),
	}

	// start the health checker asynchronously
	go func() {
		if err := h.start(); err != nil {
			log.Printf("healthcheck: start error: %v", err)
		}
	}()

	return h
}

// start begins publishing heartbeat messages and monitoring for their receipt.
// It returns an error only if initial setup (publisher/subscriber) fails.
// This method is intentionally unexported because health checking starts automatically.
func (h *HealthChecker) start() error {
	// ensure pooled connections are initialized
	if err := InitConnections(h.amqpUri); err != nil {
		return err
	}

	// create separate channels for publishing and subscribing from pooled connections
	pubConn := GetPubConn()
	if pubConn == nil {
		return amqp.ErrClosed
	}
	chPub, err := pubConn.Channel()
	if err != nil {
		return err
	}
	h.chPub = chPub

	subConn := GetSubConn()
	if subConn == nil {
		_ = h.chPub.Close()
		return amqp.ErrClosed
	}
	chSub, err := subConn.Channel()
	if err != nil {
		_ = h.chPub.Close()
		return err
	}
	h.chSub = chSub

	// declare a non-durable, auto-deleted, exclusive queue for this instance
	q, err := h.chSub.QueueDeclare(
		"",    // name
		false, // durable (non-durable)
		true,  // delete when unused (auto-delete)
		true,  // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		h.Stop()
		return err
	}

	// We'll publish directly to this queue via the default exchange (empty name)
	queueName := q.Name

	msgs, err := h.chSub.Consume(
		q.Name,
		"",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		h.Stop()
		return err
	}

	// publisher/monitor loop
	go func() {
		ticker := time.NewTicker(h.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-h.stop:
				return
			case <-ticker.C:
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				// publish directly to the queue using the default exchange
				err := h.chPub.PublishWithContext(ctx,
					"",
					queueName,
					false,
					false,
					amqp.Publishing{ContentType: "text/plain", Body: []byte("ping")},
				)
				cancel()
				if err != nil {
					log.Printf("healthcheck: publish error: %v", err)
					select {
					case h.Failure <- struct{}{}:
					default:
					}
				}

				// wait for response or timeout
				select {
				case d, ok := <-msgs:
					if !ok {
						return
					}
					_ = d
				case <-time.After(h.Timeout):
					select {
					case h.Failure <- struct{}{}:
					default:
					}
				case <-h.stop:
					return
				}
			}
		}
	}()

	return nil
}

// Stop terminates the health checker and closes underlying connections.
// It is safe to call multiple times.
func (h *HealthChecker) Stop() {
	// make close of h.stop idempotent
	defer func() { recover() }()
	close(h.stop)

	if h.chSub != nil {
		_ = h.chSub.Close()
		h.chSub = nil
	}
	if h.chPub != nil {
		_ = h.chPub.Close()
		h.chPub = nil
	}
}
