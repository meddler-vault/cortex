package consumer

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/meddler-io/watchdog/bootstrap"
	"github.com/meddler-io/watchdog/watchdog"
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
	host := getenvStr("RMQ_HOST", "192.168.29.9")
	log.Println("username", username)
	log.Println("password", password)
	log.Println("host", host)
	log.Println("MESSAGEQUEUE", bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)
	// password := getenvStr("PORt", "bitnami")

	queue := NewQueue("amqp://"+username+":"+password+"@"+host, bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)
	defer queue.Close()

	queue.Consume(func(msg string) {
		log.Println("**************************")

		defer func() {
			log.Println("**************************")
			log.Println()
		}()

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "PRE")

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
		log.Println("SystemConstants preProcess: SuccessEndpoint:", data.SuccessEndpoint)
		log.Println("SystemConstants preProcess: FailureEndpoint:", data.FailureEndpoint)

		if err = bootstrap.Bootstrap(); err != nil {
			log.Println("Error Bootstraping")
			log.Println(err)
			return

		}

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "Sync")

		for _, dependency := range data.Dependencies {
			bucketID := dependency.Identifier

			log.Println("dependency", dependency)
			if err = bootstrap.SyncStorageToDir(bucketID, *bootstrap.CONSTANTS.System.INPUTDIR, bucketID, false, true); err != nil {
				log.Println("Erro INP Sync")
				log.Println(err)
				return

			}

		}

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "Bootstrap")

		// FOrkng Process
		log.Println("Starting task")

		// environment := data.Environ

		environment := data.Environ
		if environment == nil {
			environment = make(map[string]string)

		}

		for k, v := range bootstrap.CONSTANTS.GenerateMapForSystemEnv() {
			environment[k] = v
		}

		// Replace variables & placeholders
		if data.SubstituteVariables {
			for i, arg := range data.Args {
				for k, v := range data.Variables {
					if strings.HasPrefix(v, "$") {
						if val, ok := environment[v[1:]]; ok {
							v = val
						}
					}

					data.Args[i] = strings.ReplaceAll(arg, "$"+k, v)
					arg = data.Args[i]
				}
			}
		}

		// watchdog.Start(data.Cmd, data.Args, data.Config.GenerateMapForProcessEnv())
		processErr := watchdog.Start(data.Identifier, data.Cmd, data.Args, environment)
		log.Println("Finished task", "Error:", processErr)
		// Process Finished

		log.Println("Starting OUT Sync")
		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "POST")

		if err = bootstrap.SyncDirToStorage(data.Identifier, *bootstrap.CONSTANTS.System.OUTPUTDIR, false, true); err != nil {
			log.Println("Error OUT Sync")
			log.Println(err)
			return

		}

		// Uplaod RESULTS API to Mongo

		// TODO: Inplement results DB Sync

		endpoint := ""

		headers := make(map[string]string)

		if processErr != nil {
			endpoint = data.FailureEndpoint
			headers["status"] = "false"
			headers["messsage"] = processErr.Error()

		} else {
			endpoint = data.SuccessEndpoint
			headers["status"] = "true"
			headers["messsage"] = "Successfully completed"

		}

		client := &http.Client{}
		req, _ := http.NewRequest("POST", endpoint, nil)
		for k, v := range headers {
			req.Header.Add(k, v)

		}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return
		}

		//We Read the response body on the line below.
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return
		}
		//Convert the body to type string
		sb := string(body)
		log.Println(sb)

		log.Println("Finished Sync")

	})

	// queue.Consume(func(i string) {
	// 	log.Printf("Received message with first consumer: %s", i)
	// })

	<-forever
}
