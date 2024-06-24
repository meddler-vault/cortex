package producer

import (
	"os"

	"github.com/meddler-io/watchdog/logger"
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
		logger.Fatalln(msg, err)
	}
}

func Produce(username string, password string, host string, topic string, data string) error {

	logger.Println("MQ_PUBLISH_RESULT", username, password, host, topic)
	logger.Println("MQ_PUBLISH_RESULT", data)
	conn, err := amqp.Dial("amqp://" + username + ":" + password + "@" + host)

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		topic, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	// failOnError(err, "Failed to declare a queue")
	if err != nil {
		return err
	}

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

	return err
	// failOnError(err, "Failed to publish a message")
	// log.Println(" [x] Sent %s", body)
}
