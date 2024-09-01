package consumernats

import (
	"crypto/tls"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/meddler-vault/cortex/logger"
	"github.com/meddler-vault/cortex/selfupdate"
	"github.com/nats-io/nats.go"
)

const globalTimeoutInterval = 4 * time.Second

type queue struct {
	url        string
	name       string
	topics     []string
	consumerId string

	connection   *nats.Conn
	js           nats.JetStreamContext
	subscription *nats.Subscription
	closed       bool

	currentConsumer messageConsumer

	mu sync.Mutex
}

type messageConsumer func(string, string) error

func NewQueue(url string, qName string, consumerId string, topics []string) *queue {
	reconnectInterval := 10 * time.Second
	q := new(queue)
	q.url = url
	q.name = qName
	q.consumerId = consumerId
	q.topics = topics

	// Retry to have an initla connectin foreverx
	for {
		log.Println("Atttempt to Connect", qName)

		err := q.connect()

		if err == nil {
			log.Println("Atttempt to Connect Success", qName)

			break
		}

		q.Close()
		log.Println("Will reatttempt to Connect in ", reconnectInterval, " ", qName)

		time.Sleep(reconnectInterval)
	}
	log.Println("Connect", qName)
	// go q.reconnector()

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
	if err != nil {
		logError("Error in registering consumer Consume", err)
	} else {

		logger.Println("Consumer registered! Processing messages...")

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

		if q.closed {
			logError("Reconnecting after connection closed", errors.New("connection closed"))
			q.connect()
			q.recoverConsumer()
		}
		// else if q.closed {
		// return
		// }
		time.Sleep(1 * time.Second) // Add a sleep to prevent tight loop
	}
}

func (q *queue) connect() (err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for {
		logger.Println("Connecting to NATS on ", q.url)

		// Create a custom TLS configuration
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Skip certificate verification
		}
		op := &nats.Options{
			Url:            q.url,
			ReconnectWait:  1 * time.Second,
			AllowReconnect: true,
			PingInterval:   5 * time.Second,
			MaxReconnect:   -1,
			MaxPingsOut:    1,
			// Secure:         false,     // Enable TLS
			TLSConfig: tlsConfig, // Custom TLS settings

			DisconnectedErrCB: func(nc *nats.Conn, err error) {
				logger.Println("Disconnected from NATS: ", err)
			},

			DisconnectedCB: func(nc *nats.Conn) {
				logger.Println("Disconnected from NATS: without-error")
			},
			ReconnectedCB: func(nc *nats.Conn) {
				logger.Println("Reconnected to NATS at ", nc.ConnectedUrl())
			},
			ConnectedCB: func(nc *nats.Conn) {
				logger.Println("Connected to NATS at ", nc.ConnectedUrl())
			},

			RetryOnFailedConnect: true,
			IgnoreAuthErrorAbort: true,
		}

		conn, err := op.Connect()
		if err == nil {
			q.connection = conn

			// Set up JetStream context
			js, err := q.connection.JetStream()
			if err != nil {
				logError("Error getting JetStream context", err)
				return err
			}
			q.js = js

			// Ensure the stream is durable
			log.Println("Adding Stream", q.name)

			//
			// If stream does not exist..this has to be handled by cneteal authority...just quit

			_, err = q.js.StreamInfo(q.name)
			if err != nil {
				return err
			}
			// _, err = q.js.AddStream(&nats.StreamConfig{
			// 	Name:      q.name,
			// 	Subjects:  q.topics,
			// 	Retention: nats.WorkQueuePolicy,
			// 	MaxMsgs:   -1,
			// })

			if err != nil {
				log.Println("Error adding stream..may be it already exists", err, q.topics)
				// return err
				err = nil
			}

			//

			// Ensure the stream is durable

			//

			q.closed = false

			logger.Println("Connection established!")
			return err
		} else {
			logError("Connection to NATS failed. Retrying in 1 sec... ", err)

			return err
		}

		time.Sleep(1 * time.Second) // Add a sleep to prevent tight loop
	}
}

func (q *queue) registerQueueConsumer(consumer messageConsumer) error {
	// Unsubscribe the existing consumer if it exists

	q.mu.Lock()
	defer q.mu.Unlock()
	if q.subscription != nil {

		if err := q.subscription.Unsubscribe(); err != nil {
			logError("Error unsubscribing existing consumer", err)
			return err
		}
		q.subscription = nil
	}

	// Function to subscribe to the queue

	// log.Println("Invalid log", subscribe)

	// err := subscribe(q.js)

	// New way out

	log.Println("New subscription mechanism")

	//
	// Delete the consumer that is already registered so as to handle
	//

	// Step 1: Delete the existing consumer if it exists

	// err := q.js.DeleteConsumer(q.name, q.consumerId)
	// if err != nil && !errors.Is(err, nats.ErrConsumerNotFound) {
	// 	log.Println("error deleting consumer", q.name, q.consumerId, err)
	// 	err = nil
	// } else {
	// 	log.Println("deleted already existing consumer", q.name, q.consumerId)

	// }

	//

	var sub *nats.Subscription
	for {
		var err error

		subscriptionSubject := q.topics[0]
		sub, err = q.js.QueueSubscribeSync(
			subscriptionSubject,
			q.name+"_group",
			// nats.DeliverNew(),
			// nats.MaxAckPending(1),
			nats.MaxDeliver(1),
			nats.ManualAck(), nats.Durable(q.consumerId+"-durable-consumer"))

		log.Println("Sub Sync", err, q.name, sub.IsValid(), subscriptionSubject)
		if err == nil {
			break
		}

		time.Sleep(globalTimeoutInterval)

	}

	for {

		// Checking for updates
		updateRestartReqErr := selfupdate.DoUpdateInBetweenRuntimeCheck(WatchdogVersion)

		// Try to close
		if updateRestartReqErr == nil {
			log.Println("+++++++ [[Force Restarting-Closing-NATS-Q Startup]] +++++++")
			q.Close()
			selfupdate.ForceQuit()
		} else {
			log.Println("error", updateRestartReqErr)
		}

		//
		//

		if !sub.IsValid() {
			time.Sleep(globalTimeoutInterval)
			continue
		}

		msg, err := sub.NextMsg(5 * time.Second)
		if err != nil {
			log.Println("NextMsg", err, sub.IsValid())

			time.Sleep(globalTimeoutInterval)
			continue

		}

		err = msg.AckSync()
		if err != nil {
			log.Println("NextMsg", err)

			time.Sleep(globalTimeoutInterval)
			continue

		} else {
			log.Println("Acknowledged")

		}

		logger.Println("--consumer--starts--")
		status := consumer(string(msg.Data), msg.Subject)
		logger.Println("--consumer--ends--", status)

	}

	return nil
}

func (q *queue) recoverConsumer() {
	if q.currentConsumer != nil {
		logger.Println("Recovering consumer...")

		if q.subscription != nil {
			if err := q.subscription.Unsubscribe(); err != nil {
				logError("Error unsubscribing existing consumer during recovery", err)
				// return
			}
			q.subscription = nil
		}

		err := q.registerQueueConsumer(q.currentConsumer)
		if err != nil {
			logger.Println("Error in recovering consumer", err)
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
