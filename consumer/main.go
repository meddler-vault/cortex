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

func Start() {
	forever := make(chan bool)

	username := getenvStr("RMQ_USERNAME", "user")
	password := getenvStr("RMQ_PASSWORD", "bitnami")
	host := getenvStr("RMQ_HOST", "localhost")
	// password := getenvStr("PORt", "bitnami")

	queue := NewQueue("amqp://"+username+":"+password+"@"+host, bootstrap.CONSTANTS.ReservedConstants.MESSAGEQUEUE)
	defer queue.Close()

	queue.Consume(func(msg string) {
		log.Printf("Received message with second consumer: %s", msg)

		// log.Printf(" [x] %s", msg)
		// data := make(map[string]string)
		data := &bootstrap.MessageSpec{}
		err := json.Unmarshal([]byte(msg), &data)
		if err != nil {
			log.Println(err, "Invalid data format")
			return
		}

		// log.Println("MessageSpec", data)

		// Prepare / Reset ENV & FS
		if err = bootstrap.Bootstrap(); err != nil {
			log.Println("Error Bootstraping")
			log.Println(err)
			return

		}

		log.Println("Starting INP Sync", bootstrap.CONSTANTS.SystemConstants.INPUTDIR)

		for _, dependency := range data.Dependencies {
			bucketID := dependency.Identifier

			log.Println("dependency", dependency)
			if err = bootstrap.SyncStorageToDir(bucketID, bootstrap.CONSTANTS.SystemConstants.INPUTDIR, bucketID, false, true); err != nil {
				log.Println("Erro INP Sync")
				log.Println(err)
				return

			}

		}

		// FOrkng Process
		log.Println("Starting task")
		watchdog.Start(data.Environ)
		log.Println("Finished task")
		// Process Finished

		log.Println("Starting OUT Sync", bootstrap.CONSTANTS.SystemConstants.OUTPUTDIR)

		if err = bootstrap.SyncDirToStorage(data.Identifier, bootstrap.CONSTANTS.SystemConstants.OUTPUTDIR, false, true); err != nil {
			log.Println("Error OUT Sync")
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
