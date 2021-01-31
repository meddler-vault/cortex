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

		bootstrap.CONSTANTS.Override(&data.Config)

		log.Println("SystemConstants preProcess: BASEPATH:", *bootstrap.CONSTANTS.System.BASEPATH)
		log.Println("SystemConstants preProcess: INPUTDIR:", *bootstrap.CONSTANTS.System.INPUTDIR)
		log.Println("SystemConstants preProcess: OUTPUTDIR:", *bootstrap.CONSTANTS.System.OUTPUTDIR)

		if err = bootstrap.Bootstrap(); err != nil {
			log.Println("Error Bootstraping")
			log.Println(err)
			return

		}

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

		environment := data.Environ

		for k, v := range bootstrap.CONSTANTS.GenerateMapForSystemEnv() {
			environment[k] = v
		}

		// watchdog.Start(data.Cmd, data.Args, data.Config.GenerateMapForProcessEnv())
		watchdog.Start(data.Cmd, data.Args, environment)
		log.Println("Finished task")
		// Process Finished

		log.Println("Starting OUT Sync")

		if err = bootstrap.SyncDirToStorage(data.Identifier, *bootstrap.CONSTANTS.System.OUTPUTDIR, false, true); err != nil {
			log.Println("Error OUT Sync")
			log.Println(err)
			return

		}

		log.Println("Finished Sync")
		log.Println("**************************")
		log.Println()

	})

	// queue.Consume(func(i string) {
	// 	log.Printf("Received message with first consumer: %s", i)
	// })

	<-forever
}
