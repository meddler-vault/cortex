package consumer

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/meddler-io/watchdog/logger"

	"net/http"
	"os"
	"strings"

	"github.com/meddler-io/watchdog/bootstrap"
	"github.com/meddler-io/watchdog/producer"
	"github.com/meddler-io/watchdog/watchdog"
)

func getenvStr(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func PublishEndResult(username string, password string, host string, data string) {
	// queue := NewQueue("amqp://"+username+":"+password+"@"+host, bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)
	err := producer.Produce(username, password, host, bootstrap.CONSTANTS.Reserved.PUBLISHMESSAGEQUEUE, data)
	logger.Println("TASK_RESULT_ERR", err)

}

func Start() {
	forever := make(chan bool)

	username := getenvStr("RMQ_USERNAME", "user")
	password := getenvStr("RMQ_PASSWORD", "bitnami")
	host := getenvStr("RMQ_HOST", "192.168.29.9")
	logger.Println("username", username)
	logger.Println("password", password)
	logger.Println("host", host)
	logger.Println("MESSAGEQUEUE", bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)
	// password := getenvStr("PORt", "bitnami")

	queue := NewQueue("amqp://"+username+":"+password+"@"+host, bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)
	defer queue.Close()

	queue.Consume(func(msg string) {
		logger.Println("**************************")

		defer func() {
			logger.Println("**************************")
			logger.Println()
		}()

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "PRE")

		bootstrap.CONSTANTS.Reset()
		data := &bootstrap.MessageSpec{}

		err := json.Unmarshal([]byte(msg), &data)
		if err != nil {
			logger.Println(err, "Invalid data format")
			return
		}

		bootstrap.CONSTANTS.Override(&data.Config)
		identifier := &data.Identifier

		//

		// Mark Initiated

		taskInitiated := bootstrap.TaskResult{}
		taskInitiated.Identifier = data.Identifier
		taskInitiated.WatchdogVersion = "unknown"
		taskInitiated.Status = "INITIATED"
		taskInitiated.Message = "Task Initiated"

		taskInitiatedString, err := json.Marshal(taskInitiated)
		if err != nil {
			logger.Println(err)
			taskInitiatedString = []byte{}
		}

		PublishEndResult(username, password, host, string(taskInitiatedString))

		//

		logger.InitNewTask(*identifier)

		resultsJsonPath := *bootstrap.CONSTANTS.System.RESULTSJSON

		logger.Println("SystemConstants preProcess: BASEPATH:", *bootstrap.CONSTANTS.System.BASEPATH)
		logger.Println("SystemConstants preProcess: INPUTDIR:", *bootstrap.CONSTANTS.System.INPUTDIR)
		logger.Println("SystemConstants preProcess: OUTPUTDIR:", *bootstrap.CONSTANTS.System.OUTPUTDIR)
		logger.Println("SystemConstants preProcess: RESULTSJSON:", *bootstrap.CONSTANTS.System.RESULTSJSON)
		logger.Println("SystemConstants preProcess: SuccessEndpoint:", data.SuccessEndpoint)
		logger.Println("SystemConstants preProcess: FailureEndpoint:", data.FailureEndpoint)

		if err = bootstrap.Bootstrap(); err != nil {
			logger.Println("Error Bootstraping")
			logger.Println(err)
			return

		}

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "Sync")

		for _, dependency := range data.Dependencies {
			bucketID := dependency.Identifier

			logger.Println("dependency", dependency)
			if err = bootstrap.SyncStorageToDir(bucketID, *bootstrap.CONSTANTS.System.INPUTDIR, bucketID, false, true); err != nil {
				logger.Println("Erro INP Sync")
				logger.Println(err)
				return

			}

		}

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "Bootstrap")

		// FOrkng Process
		logger.Println("Starting task")
		logger.Println("data.Variables", data.Variables)
		logger.Println("data.SubstituteVariables", data.SubstituteVariables)

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
					// arg = data.Args[i]
				}
			}

			// Replce environment variables's placeholders with values
			for i, arg := range data.Environ {
				for k, v := range data.Variables {
					if strings.HasPrefix(v, "$") {
						if val, ok := environment[v[1:]]; ok {
							v = val
						}
					}

					data.Environ[i] = strings.ReplaceAll(arg, "$"+k, v)
				}
			}

		}

		// watchdog.Start(data.Cmd, data.Args, data.Config.GenerateMapForProcessEnv())
		processErr := watchdog.Start(data.Identifier, data.Cmd, data.Args, environment)
		logger.Println("Finished task", "Error:", processErr)
		// Process Finished

		logger.Println("Starting OUT Sync")
		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "POST")

		if err = bootstrap.SyncDirToStorage(data.Identifier, *bootstrap.CONSTANTS.System.OUTPUTDIR, false, true); err != nil {
			logger.Println("Error OUT Sync")
			logger.Println(err)
			return

		}

		// Uplaod RESULTS API to Mongo

		// TODO: Inplement results DB Sync

		taskResult := bootstrap.TaskResult{}

		taskResult.Identifier = data.Identifier
		endpoint := ""

		headers := make(map[string]string)

		if processErr != nil {
			endpoint = data.FailureEndpoint
			headers["status"] = "false"
			headers["messsage"] = processErr.Error()

			taskResult.Status = "FAILURE"
			taskResult.Message = processErr.Error()

		} else {
			endpoint = data.SuccessEndpoint
			headers["status"] = "true"
			headers["messsage"] = "Successfully completed"

			taskResult.Status = "SUCCESS"
			taskResult.Message = "Successfully completed"

		}

		client := &http.Client{}

		responseContent, err := ioutil.ReadFile(resultsJsonPath)

		if err != nil {
			logger.Println("Error reading results fike", resultsJsonPath)
			responseContent = []byte{}
		}

		taskResult.Response = string(responseContent)
		taskResult.WatchdogVersion = "unknown"

		content := bytes.NewBuffer(responseContent)

		req, _ := http.NewRequest("POST", endpoint, content)

		for k, v := range headers {
			req.Header.Add(k, v)

		}
		resp, err := client.Do(req)
		if err != nil {
			logger.Println(err)
			return
		}

		//We Read the response body on the line below.
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Println(err)
			return
		}
		//Convert the body to type string
		sb := string(body)

		taskResultString, err := json.Marshal(taskResult)
		if err != nil {
			logger.Println(err)
			taskResultString = []byte{}
		}

		PublishEndResult(username, password, host, string(taskResultString))

		logger.Println("Published to messag queue")

		logger.Println(sb)

		logger.Println("Finished Sync")

	})

	// queue.Consume(func(i string) {
	// 	logger.Printf("Received message with first consumer: %s", i)
	// })

	<-forever
}
