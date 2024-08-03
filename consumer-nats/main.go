package consumernats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/meddler-vault/cortex/logger"

	"net/url"
	"os"

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

func SendMessage(queue *queue, topic string, message string) (err error) {

	logger.Println("Sending message", topic, message)
	err = queue.SendToTopic(topic, message)
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

func SendTaskUpdate(queue *queue, topic string, taskResult bootstrap.TaskResult) (err error) {

	message, err := json.Marshal(taskResult)
	if err != nil {
		logger.Println(err)
		message = []byte{}

	}

	logger.Println("Sending message", topic, string(message))
	err = SendMessage(queue, topic, string(message))
	if err != nil {
		log.Println("Error: ", "SendTaskUpdate", err)
	}
	return
}

func Start() {
	forever := make(chan bool)

	username := getenvStr("RMQ_USERNAME", "whitehat")
	password := getenvStr("RMQ_PASSWORD", "4Jy6P)$Ep@c^SenL")

	uuid := getenvStr("uuid", "uuid")

	username = url.QueryEscape(username)
	password = url.QueryEscape(password)
	host := getenvStr("RMQ_HOST", "rmq.meddler.io:443")
	logger.Println("uuid", uuid)
	logger.Println("username", username)
	logger.Println("password", password)
	logger.Println("host", host)
	logger.Println("MESSAGEQUEUE", bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)

	logger.Println("SystemConstants preProcess: BASEPATH:", *bootstrap.CONSTANTS.System.BASEPATH)

	connectionString := fmt.Sprintf("wss://%s:%s@%s", username, password, host)
	queue := NewQueue(connectionString, bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE, uuid)

	if bootstrap.CONSTANTS.Reserved.MOCKMESSAGE != "" {

		logger.Println("Publishing", "mock-message", bootstrap.CONSTANTS.Reserved.MOCKMESSAGE)

		// err := PublishMockMessage(connectionString, bootstrap.CONSTANTS.Reserved.MOCKMESSAGE)
		err := SendMessage(queue, bootstrap.CONSTANTS.Reserved.PUBLISHMESSAGEQUEUE, string(bootstrap.CONSTANTS.Reserved.MOCKMESSAGE))

		if err != nil {

			log.Println("MOCK Mode is turned on, but coudn;t publish the message. Returning to genesis", "ERror: ", err)
			return
		}
	}

	defer queue.Close()

	queue.Consume(func(msg string) {
		logger.Println("**************************")

		// logger.Println(msg)
		logger.Println("**************************")

		defer func() {
			if r := recover(); r != nil {
				logger.Println("Recovered from panic due to unhandled exception:", r)
			}
			logger.Println("**************************")
			logger.Println()

		}()

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.INPUTDIR, "PRE")

		// For each new message, reset the env state
		bootstrap.CONSTANTS.Reset()
		data := &bootstrap.MessageSpec{}

		err := json.Unmarshal([]byte(msg), &data)
		if err != nil {
			logger.Println(err, "Invalid data format:  task-deferred", msg)
			return
		}

		// Override the constants with message-spec
		bootstrap.CONSTANTS.Override(&data.Config)
		identifier := &data.Identifier

		// Mark Initiated
		SendTaskUpdate(queue, bootstrap.CONSTANTS.Reserved.PUBLISHMESSAGEQUEUE, bootstrap.TaskResult{
			Identifier:      data.Identifier,
			TaskStatus:      bootstrap.INITIATED,
			Message:         "Task Initiated",
			WatchdogVersion: WatchdogVersion,
		})

		logger.InitNewTask(*identifier)

		logger.Println("[[Watchdog]]", WatchdogVersion)

		logger.Println("SystemConstants preProcess: BASEPATH:", *bootstrap.CONSTANTS.System.BASEPATH)
		logger.Println("SystemConstants preProcess: INPUTDIR:", *bootstrap.CONSTANTS.System.INPUTDIR)
		logger.Println("SystemConstants preProcess: OUTPUTDIR:", *bootstrap.CONSTANTS.System.OUTPUTDIR)
		logger.Println("SystemConstants preProcess: MOUNT_VOLUME_PATH:", *bootstrap.CONSTANTS.System.MOUNT_VOLUME_PATH)
		logger.Println("SystemConstants preProcess: EXPORT_VOLUME_PATH:", *bootstrap.CONSTANTS.System.EXPORT_VOLUME_PATH)
		logger.Println("SystemConstants preProcess: RESULTSJSON:", *bootstrap.CONSTANTS.System.RESULTSJSON)
		logger.Println("SystemConstants preProcess: SuccessEndpoint:", data.SuccessEndpoint)
		logger.Println("SystemConstants preProcess: FailureEndpoint:", data.FailureEndpoint)

		if err = bootstrap.Bootstrap(); err != nil {
			logger.Println("Error Bootstraping")
			logger.Println(err)
			return

		}

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.BASEPATH, "Sync")

		// Mount all dependencies. Move it minio mounting later
		for _, dependency := range data.Dependencies {
			// bucketID := dependency.Identifier

			dependency.ResolveRelativePathsInDependencies(*bootstrap.CONSTANTS.System.BASEPATH)

			log.Println("Dependency-test", *dependency.MOUNT_VOLUME_PATH)
			// _, _, _ :=
			fp, fileP, err := bootstrap.SyncMountVolumedToHost(

				*dependency.MOUNT_VOLUME_S3_HOST,
				*dependency.MOUNT_VOLUME_S3_ACCESS_KEY,
				*dependency.MOUNT_VOLUME_S3_SECRET_KEY,
				*dependency.MOUNT_VOLUME_S3_SECURE,
				*dependency.MOUNT_VOLUME_S3_REGION,
				*dependency.MOUNT_VOLUME_PATH,

				*dependency.MOUNT_VOLUME_BUCKET,
				*dependency.MOUNT_VOLUME_FOLDER_PATH,
				*dependency.MOUNT_VOLUME_OBJECT_PATH,
				true,
				true,
			)

			if err != nil {
				return
			}

			data.Variables[*dependency.MOUNT_VOLUME_VARIABLE] = "$" + *dependency.MOUNT_VOLUME_VARIABLE
			data.Environ[*dependency.MOUNT_VOLUME_VARIABLE] = fp

			logger.Println("dependency-data", fp, fileP, err)

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

		// Check and populate git volume
		if *bootstrap.CONSTANTS.System.GITMODE {

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

		// Check and populate minio volume
		logger.Println("Sync minio-mount")
		if *bootstrap.CONSTANTS.System.MOUNT_VOLUME {

			logger.Println("Sync Initiate::  Mount Minio Volume", *bootstrap.CONSTANTS.System.MOUNT_VOLUME,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_BUCKET,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_FOLDER_PATH,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_OBJECT_PATH,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_PATH,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_S3_ACCESS_KEY,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_S3_SECRET_KEY,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_S3_HOST,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_S3_SECURE,
			)

			folderPath, filePath, err := bootstrap.SyncMountVolumedToHost(

				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_S3_HOST,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_S3_ACCESS_KEY,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_S3_SECRET_KEY,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_S3_SECURE,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_S3_REGION,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_PATH,

				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_BUCKET,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_FOLDER_PATH,
				*bootstrap.CONSTANTS.System.MOUNT_VOLUME_OBJECT_PATH,
				true,
				true,
			)

			if err != nil {
				logger.Println("Erro Syncing Minio Volume", err)
				return
			} else {
				logger.Println("minio-mount:success", "folderPath->", folderPath, "filePath->", filePath)
				bootstrap.CONSTANTS.System.MOUNT_VOLUME_FOLDER_PATH = &folderPath
				bootstrap.CONSTANTS.System.MOUNT_VOLUME_OBJECT_PATH = &filePath

			}

		}

		logger.Println("GIT Sync", " : ", "COMPLETED")

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

		bootstrap.PrintDir(*bootstrap.CONSTANTS.System.BASEPATH, "Basepath Bootstrap Print")

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

		logger.Println("Sync minio-export")
		if *bootstrap.CONSTANTS.System.EXPORT_VOLUME {

			logger.Println("Sync Initiate::  Mount Minio Volume", *bootstrap.CONSTANTS.System.EXPORT_VOLUME,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_BUCKET,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_FOLDER_PATH,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_OBJECT_PATH,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_PATH,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_S3_ACCESS_KEY,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_S3_SECRET_KEY,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_S3_HOST,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_S3_SECURE,
			)

			err := bootstrap.ExportDirToStorage(

				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_S3_HOST,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_S3_ACCESS_KEY,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_S3_SECRET_KEY,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_S3_SECURE,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_S3_REGION,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_PATH,

				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_BUCKET,
				*bootstrap.CONSTANTS.System.EXPORT_VOLUME_FOLDER_PATH,
				true,
				true,
			)

			if err != nil {
				logger.Println("Erro Exporting Minio Volume", err)
				return
			} else {
				logger.Println("minio-export:success")

			}

		}

		// Uplaod RESULTS API to Mongo

		// TODO: Inplement results DB Sync

		if processErr != nil {
			// Mark Finished failure
			SendTaskUpdate(queue, bootstrap.CONSTANTS.Reserved.PUBLISHMESSAGEQUEUE, bootstrap.TaskResult{
				Identifier:      data.Identifier,
				TaskStatus:      bootstrap.FAILURE,
				Message:         processErr.Error(),
				WatchdogVersion: WatchdogVersion,
			})

		} else {
			// Mark Finished success
			SendTaskUpdate(queue, bootstrap.CONSTANTS.Reserved.PUBLISHMESSAGEQUEUE, bootstrap.TaskResult{
				Identifier:      data.Identifier,
				TaskStatus:      bootstrap.SUCCESS,
				Message:         "Task completed successfully",
				WatchdogVersion: WatchdogVersion,
			})

		}

		// err = PublishEndResult(connectionString, string(taskResultString))
		// err = SendMessage(queue, bootstrap.CONSTANTS.Reserved.PUBLISHMESSAGEQUEUE, string(taskResultString))

		// Mark Finished
		SendTaskUpdate(queue, bootstrap.CONSTANTS.Reserved.PUBLISHMESSAGEQUEUE, bootstrap.TaskResult{
			Identifier:      data.Identifier,
			TaskStatus:      bootstrap.SUCCESS,
			Message:         "Task Success",
			WatchdogVersion: WatchdogVersion,
		})
		//

		logger.Println("Published to messag queue")

		logger.Println("Finished Sync")

	})

	// queue.Consume(func(i string) {
	// 	logger.Printf("Received message with first consumer: %s", i)
	// })

	<-forever
}
