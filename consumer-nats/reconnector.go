package consumernats

import (
	"log"
	"time"

	"github.com/meddler-vault/cortex/logger"
	"github.com/nats-io/nats.go"
)

type queue struct {
	url  string
	name string

	connection   *nats.Conn
	js           nats.JetStreamContext
	subscription *nats.Subscription
	closed       bool

	currentConsumer messageConsumer
}

type messageConsumer func(string)

func NewQueue(url string, qName string) *queue {
	q := new(queue)
	q.url = url
	q.name = qName

	q.connect()
	log.Println("Connect", qName)
	go q.reconnector()

	return q
}

func (q *queue) Send(message string) (err error) {
	_, err = q.js.Publish(q.name, []byte(message))
	logError("Sending message to queue failed", err)
	return
}

func (q *queue) SendToTopic(topic string, message string) (err error) {
	_, err = q.js.Publish(topic, []byte(message))
	logError("Sending message to queue failed", err)
	return
}

func (q *queue) Consume(consumer messageConsumer) {
	logger.Println("Registering consumer...")
	err := q.registerQueueConsumer(consumer)
	logger.Println("Consumer registered! Processing messages...")
	if err != nil {
		logError("Error in registering consumer", err)
	}
}

func (q *queue) Close() {
	logger.Println("Closing connection")
	q.closed = true
	if q.subscription != nil {
		q.subscription.Unsubscribe()
	}
	q.connection.Close()
}

func (q *queue) reconnector() {
	for {
		if q.connection.IsClosed() && !q.closed {
			logError("Reconnecting after connection closed", nil)
			q.connect()
			q.recoverConsumer()
		} else if q.closed {
			return
		}
		time.Sleep(1 * time.Second) // Add a sleep to prevent tight loop
	}
}

func (q *queue) connect() {
	for {
		logger.Println("Connecting to NATS on ", q.url)

		op := &nats.Options{
			Url:           q.url,
			ReconnectWait: 1 * time.Second,
			PingInterval:  5 * time.Second,
			MaxReconnect:  1,
			MaxPingsOut:   1,
		}

		conn, err := op.Connect()
		if err == nil {
			q.connection = conn

			// Set up JetStream context
			js, err := q.connection.JetStream()
			if err != nil {
				logError("Error getting JetStream context", err)
				return
			}
			q.js = js

			// Ensure the stream is durable
			_, err = q.js.AddStream(&nats.StreamConfig{
				Name:     q.name,
				Subjects: []string{q.name},
				Storage:  nats.FileStorage,
			})
			if err != nil {
				logError("Error adding stream", err)
				return
			}

			logger.Println("Connection established!")
			return
		}

		logError("Connection to NATS failed. Retrying in 1 sec... ", err)
		time.Sleep(1 * time.Second) // Add a sleep to prevent tight loop
	}
}

func (q *queue) registerQueueConsumer(consumer messageConsumer) error {
	// Unsubscribe the existing consumer if it exists
	if q.subscription != nil {
		q.subscription.Unsubscribe()
	}

	sub, err := q.js.Subscribe(q.name, func(msg *nats.Msg) {
		consumer(string(msg.Data))
	}, nats.Durable("durable-consumer"))
	if err == nil {
		q.subscription = sub
		q.currentConsumer = consumer
	}
	return err
}

func (q *queue) recoverConsumer() {
	if q.currentConsumer != nil {
		logger.Println("Recovering consumer...")
		err := q.registerQueueConsumer(q.currentConsumer)
		if err != nil {
			logError("Error in recovering consumer", err)
		} else {
			logger.Println("Consumer recovered! Continuing message processing...")
		}
	}
}

func logError(message string, err error) {
	if err != nil {
		logger.Println(message, err)
	}
}
