package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/meddler-xyz/watchdog/consumer"

	"github.com/meddler-xyz/watchdog/watchdog"
	"github.com/streadway/amqp"
)

var amqpUri = flag.String("RabbitMQ_URI", "amqp://user:bitnami@127.0.0.1", "RabbitMQ URI")

var (
	rabbitConn       *amqp.Connection
	rabbitCloseError chan *amqp.Error
)

// Try to connect to the RabbitMQ server as
// long as it takes to establish a connection
//
func connectToRabbitMQ(uri string) *amqp.Connection {
	for {
		conn, err := amqp.Dial(uri)

		if err == nil {
			return conn
		}

		log.Println(err)
		log.Printf("Trying to reconnect to RabbitMQ at %s\n", uri)
		time.Sleep(500 * time.Millisecond)
	}
}

// re-establish the connection to RabbitMQ in case
// the connection has died
//
func rabbitConnector(uri string) {
	var rabbitErr *amqp.Error

	for {
		rabbitErr = <-rabbitCloseError
		if rabbitErr != nil {
			log.Printf("Connecting to %s\n", *amqpUri)

			rabbitConn = connectToRabbitMQ(uri)
			rabbitCloseError = make(chan *amqp.Error)
			rabbitConn.NotifyClose(rabbitCloseError)

			setupWorker(rabbitConn)
			// run your setup process here
		}
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

func setupWorker(conn *amqp.Connection) {
	fmt.Println("Seting up worker")

	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"tasks", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	failOnError(err, "Could not create queue")
	err = ch.Qos(1, 0, false)
	failOnError(err, "Could not configure QOS")

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto ack
		false,      // exclusive
		false,      // no local
		false,      // no wait
		nil,        // args
	)

	failOnError(err, "Could not register consumer ")

	for msg := range msgs {
		log.Printf(" [x] %s", msg.RoutingKey)
		data := make(map[string]string)
		err = json.Unmarshal(msg.Body, &data)
		if err != nil {
			failOnError(err, "Invalid format")
			if err := msg.Ack(false); err != nil {
				failOnError(err, "Failed to ack")
			} else {

				log.Println("Successfully ackd")

			}
			break

		}

		log.Println("Starting task")
		watchdog.Start(data)
		log.Println("Finished task")
		if err := msg.Ack(false); err != nil {
			failOnError(err, "Failed to ack")
		} else {
			log.Println("Successfully ackd")

		}
	}

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")

}

func main() {

	consumer.Main()
}
func a2main() {
	forever := make(chan bool)

	flag.Parse()

	// create the rabbitmq error channel
	rabbitCloseError = make(chan *amqp.Error)

	// run the callback in a separate thread
	log.Println("Conneting to ", *amqpUri)
	go rabbitConnector(*amqpUri)

	// establish the rabbitmq connection by sending
	// an error and thus calling the error callback
	rabbitCloseError <- amqp.ErrClosed
	<-forever

}
