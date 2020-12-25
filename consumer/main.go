package consumer

import (
	"encoding/json"
	"log"

	"github.com/meddler-xyz/watchdog/watchdog"
)

func Main() {
	forever := make(chan bool)

	queue := NewQueue("amqp://user:bitnami@127.0.0.1", "tasks")
	defer queue.Close()

	queue.Consume(func(msg string) {
		log.Printf("Received message with second consumer: %s", msg)

		log.Printf(" [x] %s", msg)
		data := make(map[string]string)
		err := json.Unmarshal([]byte(msg), &data)
		if err != nil {
			log.Println(err, "Invalid format")

			// if err := msg.Ack(false); err != nil {
			// 	log.Println(err, "Failed to ack")
			// } else {

			// 	log.Println("Successfully ackd")

			// }

		}

		log.Println("Starting task")
		watchdog.Start(data)
		log.Println("Finished task")

	})

	// queue.Consume(func(i string) {
	// 	log.Printf("Received message with first consumer: %s", i)
	// })

	<-forever
}
