package producer

import (
	"log"
	"os"

	"github.com/meddler-io/watchdog/bootstrap"
	"github.com/streadway/amqp"
)

func getenvStr(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func Produce(data string) {

	username := getenvStr("RMQ_USERNAME", "user")
	password := getenvStr("RMQ_PASSWORD", "bitnami")
	host := getenvStr("RMQ_HOST", "localhost")
	// password := getenvStr("PORt", "bitnami")

	conn, err := amqp.Dial("amqp://" + username + ":" + password + "@" + host)

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := data
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	failOnError(err, "Failed to publish a message")
	// log.Printf(" [x] Sent %s", body)
}
