package consumernats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/meddler-vault/cortex/logger"

	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/meddler-vault/cortex/bootstrap"
	"github.com/meddler-vault/cortex/watchdog"
)

var WatchdogVersion = "version"

func getenvStr(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func PublishEndResult(connectionString, message string) (err error) {

	queue := NewQueue(connectionString, bootstrap.CONSTANTS.Reserved.PUBLISHMESSAGEQUEUE)

	err = queue.Send(message)
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	err = queue.connection.FlushWithContext(
		ctx,
	)
	if err != nil {
		return err
	}

	return

}

func PublishMockMessage(connectionString string, message string) (err error) {

	queue := NewQueue(connectionString, bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)

	err = queue.Send(message)
	if err != nil {
		return err
	}

	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	err = queue.connection.FlushWithContext(
		ctx,
	)
	if err != nil {
		return err
	}

	// queue.connection.Close()

	return

}

func Start() {
	forever := make(chan bool)

	username := getenvStr("RMQ_USERNAME", "whitehat")
	password := getenvStr("RMQ_PASSWORD", "4Jy6P)$Ep@c^SenL")
	username = url.QueryEscape(username)
	password = url.QueryEscape(password)
	host := getenvStr("RMQ_HOST", "rmq.meddler.io:443")
	logger.Println("username", username)
	logger.Println("password", password)
	logger.Println("host", host)
	logger.Println("MESSAGEQUEUE", bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)

	connectionString := fmt.Sprintf("wss://%s:%s@%s", username, password, host)

	if bootstrap.CONSTANTS.Reserved.MOCKMESSAGE != "" {

		logger.Println("MOCKMESSAGE", bootstrap.CONSTANTS.Reserved.MOCKMESSAGE)
		err := PublishMockMessage(connectionString, bootstrap.CONSTANTS.Reserved.MOCKMESSAGE)
		if err != nil {

			log.Println("MOCK Mode is turned on, but coudn;t publish the message. Returning to genesis", "ERror: ", err)
			return
		}
	}

	queue := NewQueue(connectionString, bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)

	defer queue.Close()

	queue.Consume(func(msg string) {
		logger.Println("**************************")
		// logger.Println(msg)
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
		taskInitiated.WatchdogVersion = WatchdogVersion
		taskInitiated.Status = "INITIATED"
		taskInitiated.Message = "Task Initiated"

		taskInitiatedString, err := json.Marshal(taskInitiated)
		if err != nil {
			logger.Println(err)
			taskInitiatedString = []byte{}
		}

		PublishEndResult(connectionString, string(taskInitiatedString))

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

		// environment := data.Environ

		environment := data.Environ
		if environment == nil {
			environment = make(map[string]string)

		}

		for k, v := range bootstrap.CONSTANTS.GenerateMapForSystemEnv() {
			environment[k] = v
		}

		logger.Println("**environment.environment**", environment)
		// Replace variables & placeholders
		if data.SubstituteVariables {

			// Replce environment variables's placeholders with values
			logger.Println("**SubstituteVariables**", "Environ", "before", data.Environ)

			for i, arg := range data.Environ {
				for k, v := range data.Variables {
					if strings.HasPrefix(v, "$") {
						if val, ok := environment[v[1:]]; ok {
							v = val
						}
					}

					arg = strings.ReplaceAll(arg, "$"+k, v)

					data.Environ[i] = arg
				}
			}

			logger.Println("**SubstituteVariables**", "Environ", "after", data.Environ)

			logger.Println("**SubstituteVariables**")
			logger.Println("**SubstituteVariables**", "Entrypoint", "before", data.Entrypoint)

			// Placoholder replacment for entrypoint
			for i, arg := range data.Entrypoint {
				for k, v := range data.Variables {
					if strings.HasPrefix(v, "$") {
						if val, ok := environment[v[1:]]; ok {
							v = val
						}
					}

					arg = strings.ReplaceAll(arg, "$"+k, v)
					data.Entrypoint[i] = arg
					// arg = data.Args[i]
				}
			}
			logger.Println("**SubstituteVariables**", "Entrypoint", "after", data.Entrypoint)

			// Placoholder replacment for cmd
			logger.Println("**SubstituteVariables**", "Cmd", "before", data.Cmd)

			for i, arg := range data.Cmd {
				for k, v := range data.Variables {
					if strings.HasPrefix(v, "$") {
						if val, ok := environment[v[1:]]; ok {
							v = val
						}
					}

					arg = strings.ReplaceAll(arg, "$"+k, v)
					data.Cmd[i] = arg
					// arg = data.Args[i]
				}
			}
			logger.Println("**SubstituteVariables**", "Cmd", "after", data.Cmd)

			// Placeholder replacement for args

			logger.Println("**SubstituteVariables**", "Args", "before", data.Args, "environment", environment, "Variables", data.Variables)

			logger.Println("")
			logger.Println("")
			logger.Println("")
			for i, arg := range data.Args {
				logger.Println("Arg", i, arg)
				for k, v := range data.Variables {
					logger.Println("-> Variable", k, v)

					if strings.HasPrefix(v, "$") {

						if val, ok := environment[v[1:]]; ok {

							logger.Println("--> Variable : HasPrefix", val)

							v = val
						}
					}

					arg = strings.ReplaceAll(arg, "$"+k, v)
					data.Args[i] = arg

					logger.Println("==> Arg : HasPrefix", arg, "$"+k, v, data.Args[i])
					// arg = data.Args[i]
				}
			}
			logger.Println("")
			logger.Println("")
			logger.Println("")

			logger.Println("**SubstituteVariables**", "Args", "after", data.Args)

		}

		data.Cmd = append(data.Entrypoint, data.Cmd...)

		logger.Println("Sync Deps Done")
		// Load git repo locally
		logger.Println("Sync Initiate::  Git Repo",
			*bootstrap.CONSTANTS.System.GITMODE,
			*bootstrap.CONSTANTS.System.GITREMOTE,
			*bootstrap.CONSTANTS.System.GITPATH,
			*bootstrap.CONSTANTS.System.GITAUTHMODE,
			*bootstrap.CONSTANTS.System.GITAUTHUSERNAME,
			*bootstrap.CONSTANTS.System.GITAUTHPASSWORD,
			*bootstrap.CONSTANTS.System.GITREF,
			*bootstrap.CONSTANTS.System.GITDEPTH,
		)

		if strings.ToLower(*bootstrap.CONSTANTS.System.GITMODE) == "true" {

			logger.Println("Sync Initiate::  Git Repo", *bootstrap.CONSTANTS.System.GITREMOTE,
				*bootstrap.CONSTANTS.System.GITPATH,
				*bootstrap.CONSTANTS.System.GITAUTHMODE,
				*bootstrap.CONSTANTS.System.GITAUTHUSERNAME,
				*bootstrap.CONSTANTS.System.GITAUTHPASSWORD,
				*bootstrap.CONSTANTS.System.GITREF,
				*bootstrap.CONSTANTS.System.GITDEPTH,
			)

			repository, err := bootstrap.Clone(
				*bootstrap.CONSTANTS.System.GITREMOTE,
				*bootstrap.CONSTANTS.System.GITPATH,
				*bootstrap.CONSTANTS.System.GITAUTHMODE,
				*bootstrap.CONSTANTS.System.GITAUTHUSERNAME,
				*bootstrap.CONSTANTS.System.GITAUTHPASSWORD,
				*bootstrap.CONSTANTS.System.GITREF,
				*bootstrap.CONSTANTS.System.GITDEPTH,
			)

			if err != nil {
				logger.Println("Erro Syncing Git Repo", err)
				return
			} else {
				logger.Println("Finished Syncing Git Repo", repository)

			}
		}

		logger.Println("GIT Sync", " : ", "COMPLETED")

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "Bootstrap")

		// FOrkng Process
		logger.Println("Starting task")
		logger.Println("data.Variables", data.Variables)
		logger.Println("data.SubstituteVariables", data.SubstituteVariables)
		logger.Println("Reaper", data.Identifier, data.Cmd, data.Args, environment)

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

		responseContent, err := os.ReadFile(resultsJsonPath)

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
		body, err := io.ReadAll(resp.Body)
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

		PublishEndResult(connectionString, string(taskResultString))

		logger.Println("Published to messag queue")

		logger.Println(sb)

		logger.Println("Finished Sync")

	})

	// queue.Consume(func(i string) {
	// 	logger.Printf("Received message with first consumer: %s", i)
	// })

	<-forever
}
