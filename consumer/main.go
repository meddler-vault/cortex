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

	queue := NewQueue("amqp://"+username+":"+password+"@"+host, bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)
	defer queue.Close()

	queue.Consume(func(msg string) {
		log.Println("**************************")

		bootstrap.CONSTANTS.Reset()
		data := &bootstrap.MessageSpec{}

		err := json.Unmarshal([]byte(msg), &data)
		if err != nil {
			log.Println(err, "Invalid data format")
			return
		}
		log.Printf("Received message with second consumer: %s", msg)

		// log.Printf(" [x] %s", msg)
		// data := make(map[string]string)

		log.Println("Constants via API: ", data.Config)
		// log.Println("ProcessConstants preProcess: ", bootstrap.CONSTANTS.Process)
		log.Println("SystemConstants preProcess: ", *bootstrap.CONSTANTS.System.BASEPATH)
		// log.Println("ReservedConstants preProcess: ", bootstrap.CONSTANTS.Reserved)

		bootstrap.CONSTANTS.Override(&data.Config)

		// bootstrap.CONSTANTS.Process = data.Config.Process
		// bootstrap.CONSTANTS.Reserved = data.Config.Reserved
		// bootstrap.CONSTANTS.System = data.Config.System
		// log.Println("ProcessConstants postProcess: ", bootstrap.CONSTANTS.Process)
		log.Println("SystemConstants postProcess: ", *bootstrap.CONSTANTS.System.BASEPATH)
		// log.Println("ReservedConstants postProcess: ", bootstrap.CONSTANTS.Reserved)

		// log.Println("MessageSpec", data)

		// Prepare / Reset ENV & FS
		if err = bootstrap.Bootstrap(); err != nil {
			log.Println("Error Bootstraping")
			log.Println(err)
			return

		}

		log.Println("Starting INP Sync", bootstrap.CONSTANTS.System.INPUTDIR)

		for _, dependency := range data.Dependencies {
			bucketID := dependency.Identifier

			log.Println("dependency", dependency)
			if err = bootstrap.SyncStorageToDir(bucketID, *bootstrap.CONSTANTS.System.INPUTDIR, bucketID, false, true); err != nil {
				log.Println("Erro INP Sync")
				log.Println(err)
				return

			}

		}

		// FOrkng Process
		log.Println("Starting task")
		watchdog.Start(data.Config.GenerateMapForProcessEnv())
		log.Println("Finished task")
		// Process Finished

		log.Println("Starting OUT Sync", bootstrap.CONSTANTS.System.OUTPUTDIR)

		if err = bootstrap.SyncDirToStorage(data.Identifier, *bootstrap.CONSTANTS.System.OUTPUTDIR, false, true); err != nil {
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
