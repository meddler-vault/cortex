package consumer

import (
	"encoding/json"
	"log"
	"os"

	"github.com/meddler-xyz/watchdog/watchdog"
)

func getenvStr(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func Main() {
	forever := make(chan bool)

	username := getenvStr("RMQ_USERNAME", "user")
	password := getenvStr("RMQ_PASSWORD", "bitnami")
	host := getenvStr("RMQ_HOST", "localhost")
	// password := getenvStr("PORt", "bitnami")

	queue := NewQueue("amqp://"+username+":"+password+"@"+host, "tasks")
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
