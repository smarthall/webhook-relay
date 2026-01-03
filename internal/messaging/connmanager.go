package messaging

import (
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	initOnce sync.Once
	initErr  error

	pubConn *amqp.Connection
	subConn *amqp.Connection
	mu      sync.Mutex
)

// InitConnections dials two separate AMQP connections (one for publishing,
// one for subscribing). It is safe to call multiple times; initialization
// happens only once.
func InitConnections(amqpUri string) error {
	initOnce.Do(func() {
		// publisher connection
		c1, err := amqp.Dial(amqpUri)
		if err != nil {
			initErr = err
			return
		}
		log.Printf("Connected to RabbitMQ (publisher)")

		// subscriber connection
		c2, err := amqp.Dial(amqpUri)
		if err != nil {
			// close first if second fails
			_ = c1.Close()
			initErr = err
			return
		}
		log.Printf("Connected to RabbitMQ (subscriber)")

		mu.Lock()
		pubConn = c1
		subConn = c2
		mu.Unlock()
	})
	return initErr
}

func GetPubConn() *amqp.Connection {
	mu.Lock()
	defer mu.Unlock()
	return pubConn
}

func GetSubConn() *amqp.Connection {
	mu.Lock()
	defer mu.Unlock()
	return subConn
}

// CloseConnections closes both pooled connections. It is safe to call
// multiple times.
func CloseConnections() {
	mu.Lock()
	defer mu.Unlock()
	if pubConn != nil {
		_ = pubConn.Close()
		pubConn = nil
	}
	if subConn != nil {
		_ = subConn.Close()
		subConn = nil
	}
}
