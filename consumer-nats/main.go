package consumernats

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/meddler-vault/cortex/healthchecker"
	"github.com/meddler-vault/cortex/logger"

	"os"

	"github.com/meddler-vault/cortex/bootstrap"
)

var WatchdogVersion = "0.0.1"

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

func SendTaskUpdate(queue *queue, taskResult bootstrap.TaskResult) (err error) {

	//
	healthCheckMessage := map[string]interface{}{
		"identifier":  taskResult.Identifier,
		"exec_status": taskResult.TaskStatus,
		"state":       "working", "details": "Inside working mode!"}

	healthchecker.SetMessage(healthCheckMessage)

	//
	taskResult.WorkerId = queue.workerId

	//
	message, err := json.Marshal(taskResult)
	if err != nil {
		logger.Println(err)
		message = []byte{}

	}

	logger.Println("Sending message", queue.publisherSubject, string(message))
	err = SendMessage(queue, queue.publisherSubject, string(message))
	if err != nil {
		log.Println("Error: ", "SendTaskUpdate", err)
	}
	return
}

func Start() {
	forever := make(chan bool)

	// username := getenvStr("RMQ_USERNAME", "whitehat")
	// password := getenvStr("RMQ_PASSWORD", "4Jy6P)$Ep@c^SenL")

	connectionString := getenvStr("NATS_CONNECTION_STRING", "nats://connection-string")

	// username = url.QueryEscape(username)
	// password = url.QueryEscape(password)
	// host := getenvStr("RMQ_HOST", "hawki-rabbitmq.indiatimes.com:4222")

	// logger.Println("username", username)
	// logger.Println("password", password)
	// logger.Println("host", host)
	logger.Println("MESSAGEQUEUE", bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE)
	logger.Println("natsConnString", connectionString)

	logger.Println("SystemConstants preProcess: BASEPATH:", *bootstrap.CONSTANTS.System.BASEPATH)

	// connectionString := fmt.Sprintf("nats://%s:%s@%s", username, password, host)

	// connectionString = "wss://_whitehat:4Jy6P%29%24Ep%40c%5ESenL@rmq.meddler.io:443"
	// connectionString = natsConnString

	log.Println("CORTEXMODE", bootstrap.CONSTANTS.Reserved.CORTEXMODE)

	var publisherSubject string
	var consumerSubject string
	var workerGroupName string
	var cortexMode bootstrap.CortexMode

	var msgHandler messageConsumer

	if bootstrap.CONSTANTS.Reserved.CORTEXMODE == bootstrap.CortexModeTaskWorker {

		publisherSubject = bootstrap.CORTEX_MQ_CONSUMER_SUBJECT
		consumerSubject = bootstrap.CORTEX_MQ_PUBLISHER_SUBJECT + "." + bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE
		workerGroupName = bootstrap.CORTEX_MQ_CONSUMER_NAME + "-" + bootstrap.CONSTANTS.Reserved.MESSAGEQUEUE
		cortexMode = bootstrap.CONSTANTS.Reserved.CORTEXMODE

		msgHandler = msgHandlerForTaskWorker

	} else if bootstrap.CONSTANTS.Reserved.CORTEXMODE == bootstrap.CortexModeImageBuilder {

		consumerSubject = bootstrap.CORTEX_MQ_CONSUMER_SUBJECT
		publisherSubject = bootstrap.CORTEX_MQ_PUBLISHER_SUBJECT + "." + "build"
		workerGroupName = bootstrap.CORTEX_MQ_CONSUMER_NAME
		cortexMode = bootstrap.CONSTANTS.Reserved.CORTEXMODE

		msgHandler = msgHandlerForTaskWorker

	} else if bootstrap.CONSTANTS.Reserved.CORTEXMODE == bootstrap.CortexModeTaskResultProcessor {

		consumerSubject = bootstrap.CORTEX_MQ_CONSUMER_SUBJECT
		publisherSubject = bootstrap.CORTEX_MQ_PUBLISHER_SUBJECT
		workerGroupName = bootstrap.CORTEX_MQ_CONSUMER_NAME
		cortexMode = bootstrap.CONSTANTS.Reserved.CORTEXMODE

		msgHandler = msgHandlerForTaskResultProcessor

	} else if bootstrap.CONSTANTS.Reserved.CORTEXMODE == bootstrap.CortexModeImageBuilderResultProcessor {

		consumerSubject = bootstrap.CORTEX_MQ_CONSUMER_SUBJECT
		publisherSubject = bootstrap.CORTEX_MQ_PUBLISHER_SUBJECT
		workerGroupName = bootstrap.CORTEX_MQ_CONSUMER_NAME
		cortexMode = bootstrap.CONSTANTS.Reserved.CORTEXMODE

		msgHandler = msgHandlerForTaskResultProcessor

	} else if bootstrap.CONSTANTS.Reserved.CORTEXMODE == bootstrap.CortexModeResultProcessor {

		consumerSubject = bootstrap.CORTEX_MQ_CONSUMER_SUBJECT
		publisherSubject = bootstrap.CORTEX_MQ_PUBLISHER_SUBJECT
		workerGroupName = bootstrap.CORTEX_MQ_CONSUMER_NAME
		cortexMode = bootstrap.CONSTANTS.Reserved.CORTEXMODE

		msgHandler = msgHandlerForTaskResultProcessor

	} else {

	}

	log.Println("\n debug-details \n",
		"\n {cortex-mode} \n",
		cortexMode,
		"\n {CORTEX_MQ_CONSUMER_SUBJECT} \n",
		consumerSubject,
		"\n {CORTEX_MQ_PUBLISHER_SUBJECT} \n",
		publisherSubject,
		"\n {CORTEX_MQ_WORKER_NAME} \n",
		workerGroupName,
		"\n {CORTEXUUID} \n",
		bootstrap.CONSTANTS.Reserved.CORTEXUUID,
	)

	// <-forever

	Queue := NewQueue(
		connectionString,
		workerGroupName,
		publisherSubject,
		consumerSubject,
		bootstrap.CONSTANTS.Reserved.CORTEXUUID,
	)

	defer Queue.Close()

	Queue.Consume(

		msgHandler,
	)

	<-forever

}
