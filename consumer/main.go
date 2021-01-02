package consumer

import (
	"encoding/json"
	"log"
	"os"

	"github.com/meddler-xyz/watchdog/bootstrap"
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
			log.Println(err, "Invalid data format")
			return
		}

		var bucketID string
		var bucketIDK bool
		if bucketID, bucketIDK = data["_id"]; !bucketIDK {
			//do something
			log.Println("Bucket Key '_id' not present in data")
			return
		}

		log.Println("Starting Bootstraping")

		if err = bootstrap.Bootstrap(); err != nil {
			log.Println("Error Bootstraping")
			log.Println(err)
			return

		}

		// FOrkng Process
		log.Println("Starting task")
		watchdog.Start(data)
		log.Println("Finished task")
		// Process Finished

		log.Println("Starting Sync", *bootstrap.OUTPUTDIR)
		if err = bootstrap.SyncDirToStorage(bucketID, *bootstrap.OUTPUTDIR, false, true); err != nil {
			log.Println("Error Sync")
			log.Println(err)
			return

		}

		log.Println("Finished Sync")

	})

	// queue.Consume(func(i string) {
	// 	log.Printf("Received message with first consumer: %s", i)
	// })

	<-forever
}