package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/meddler-xyz/watchdog/watchdog"

	kaf "github.com/segmentio/kafka-go"
)

func main() {
	log.Println("Starting watchdog")

	kafkaAddress := os.Getenv("KAFKA_ADDRESS")
	if kafkaAddress == "" {
		kafkaAddress = "localhost:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "test_topic"
	}

	// make a new reader that consumes from topic-A
	r := kaf.NewReader(kaf.ReaderConfig{
		Brokers:  []string{kafkaAddress},
		GroupID:  "consumer-group-id",
		Topic:    kafkaTopic,
		MinBytes: 10e0, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	log.Println("Starting listening to message queue")
	log.Println("kafkaAddress", kafkaAddress)
	log.Println("kafkaTopic", kafkaTopic)

	for {
		msg, err := r.ReadMessage(context.Background())
		fmt.Printf("Message on %s: %s\n", string(msg.Key), string(msg.Value))

		if err == nil {

			data := make(map[string]string)
			json.Unmarshal(msg.Value, &data)

			fmt.Printf("Message on %s: %s\n", string(msg.Key), data)

			watchdog.Start(data)
		} else {
			// The client will automatically try to recover from all errors.
			fmt.Printf("Consumer error: %v (%v)\n", err, msg)
		}

	}

}
