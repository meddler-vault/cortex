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
	streamName string

	publisherSubject string
	consumerSubject  string
	workerGroupName  string

	//
	url string
	// name       string
	// subject    string
	// consumerId string

	connection   *nats.Conn
	js           nats.JetStreamContext
	subscription *nats.Subscription
	closed       bool

	currentConsumer messageConsumer

	mu sync.Mutex
}

type messageConsumer func(*queue, string, string) error

func NewQueue(
	url string,
	workerGroupName string,
	publisherSubject string,
	consumerSubject string,

) *queue {
	//
	reconnectInterval := 10 * time.Second
	q := new(queue)
	q.streamName = "TASKS"

	q.url = url
	q.publisherSubject = publisherSubject
	q.consumerSubject = consumerSubject
	q.workerGroupName = workerGroupName

	// Retry to have an initla connectin foreverx
	for {
		log.Println("Atttempt to Connect")

		err := q.connect()

		if err == nil {
			log.Println("Atttempt to Connect Success")

			break
		}

		q.Close()
		log.Println("Will reatttempt to Connect in ", reconnectInterval, "duration")

		time.Sleep(reconnectInterval)
	}

	// go q.reconnector()

	return q
}

func (q *queue) SendToTopic(topic string, message string) (err error) {

	_, err = q.js.Publish(topic, []byte(message))
	log.Println("Sending message to queue failed", q.streamName, q.workerGroupName, topic, message, err)
	return
}

func (q *queue) Consume(

	consumer messageConsumer) {

	for {

		logger.Println("Registering consumer...", q.consumerSubject,

			q.consumerSubject,
			q.workerGroupName,
		)
		err := q.registerQueueConsumer(consumer)
		if err != nil {
			logError("Error in registering consumer Consume", err)
			// return
		} else {

			logger.Println("Consumer registered! Processing messages...")

		}
		time.Sleep(5 * time.Second)
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

func (q *queue) reconnector(subjectPrefix string) {
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

			//
			// If stream does not exist..this has to be handled by cneteal authority...just quit

			_, err = q.js.StreamInfo(q.streamName)
			if err != nil {
				log.Println(err)
				return err
			}
			// _, err = q.js.AddStream(&nats.StreamConfig{
			// 	Name:      q.name,
			// 	Subjects:  q.topics,
			// 	Retention: nats.WorkQueuePolicy,
			// 	MaxMsgs:   -1,
			// })

			// if err != nil {
			// 	log.Println("Error adding stream..may be it already exists", err, q.topics)
			// 	// return err
			// 	err = nil
			// }

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

	var sub *nats.Subscription

	for {
		var err error

		log.Println("subscribing_via_consumer_name", q.workerGroupName)

		sub, err = q.js.QueueSubscribeSync(
			q.consumerSubject,
			q.workerGroupName,
			nats.MaxAckPending(1),
			nats.MaxDeliver(1),
			nats.ManualAck(),
			// nats.Durable(q.consumerId+"-durable-consumer"),
		)

		log.Println("error", err)
		log.Println("Sub Sync", err)
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

			_, err := q.js.StreamInfo(q.streamName)
			if err != nil {
				log.Println("stream_error", err)
				return err

			}

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
		status := consumer(q, string(msg.Data), msg.Subject)
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
