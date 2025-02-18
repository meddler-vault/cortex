package healthchecker

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/meddler-vault/cortex/bootstrap"
	pubsub "github.com/meddler-vault/cortex/pubsub"
)

// HealthData represents the data structure for the health check payload
type HealthData struct {
	Message   map[string]interface{} `json:"message"`
	Timestamp string                 `json:"timestamp"`
}

// AtomicHealth holds the atomic health status and message
var (
	globalHealth *AtomicHealth
	pubsubref    *pubsub.PubSub

	endpoint           string
	uuid               string
	subscription_topic string
)

var created_at string
var sequence_number int = 0

func init() {
	created_at = time.Now().Format(time.RFC3339)

}

// AtomicHealth holds the atomic health status and message
type AtomicHealth struct {
	message atomic.Value
}

// SetMessage sets the atomic message
func SetMessage(newMessage map[string]interface{}) {

	// seq should increase only when a new message is set!
	sequence_number += 1
	//
	newMessage["worker_id"] = uuid
	newMessage["subscription_topic"] = subscription_topic
	newMessage["created_at"] = created_at
	newMessage["updated_at"] = time.Now().Format(time.RFC3339)
	globalHealth.message.Store(newMessage)

}

// GetMessage returns the atomic message
func (ah *AtomicHealth) GetMessage() map[string]interface{} {
	val := ah.message.Load()
	var _val map[string]interface{}

	if val == nil {
		_val = make(map[string]interface{})
	} else {
		_val = val.(map[string]interface{})
	}

	_val["sequence_number"] = sequence_number

	return _val
}

// InitializeGlobalHealth initializes the global AtomicHealth instance
func InitializeGlobalHealth(worker_id string, subscription_topic_id string, current_endpoint string, initialMessage map[string]interface{}) *pubsub.PubSub {

	if pubsubref != nil {
		log.Println("Pubsub for healthcheck already initialized!")
	} else {
		pubsubref = pubsub.NewPubSub("healthcheck-process-killer")
	}

	if globalHealth != nil {
		log.Println("Global health already initialized")
		return pubsubref
	}

	endpoint = current_endpoint
	uuid = worker_id
	subscription_topic = subscription_topic_id
	globalHealth = &AtomicHealth{}
	SetMessage(initialMessage)

	TriggerImmediateHealthCheck()

	stopCh := make(chan struct{})

	interval := time.Duration(bootstrap.CONSTANTS.Reserved.CORTEXPINGINTERVAL) * time.Second

	HealthCheckWorker(interval, stopCh)

	return pubsubref

}

// sendHealthData sends the current health data to the server
func sendHealthData(endpoint string) {
	// Prepare the latest health data with the current timestamp
	healthData := HealthData{
		Message:   globalHealth.GetMessage(),
		Timestamp: time.Now().Format(time.RFC3339), // Get the current time as a string
	}

	// Marshal the health data into JSON
	payload, err := json.Marshal(healthData)
	if err != nil {
		log.Printf("Error marshalling health data: %v", err)
		return
	}

	// Create a new POST request
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the POST request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending health check request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode == http.StatusOK {

	} else if resp.StatusCode == http.StatusResetContent {

		if val, exists := healthData.Message["identifier"]; exists {
			if identifier, ok := val.(string); ok {
				pubsubref.Publish((identifier))

			}

		}

	} else {

	}

	log.Printf("HTTP response: %s", resp.Status)

}

// HealthCheckWorker sends the latest health data to the server periodically
func HealthCheckWorker(interval time.Duration, stopCh <-chan struct{}) {

	if interval < 5 {
		interval = 5
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				log.Println("Stopping health checker...")
				return
			case <-ticker.C:
				func() {
					// Recover from panic to keep the worker running
					defer func() {
						if r := recover(); r != nil {
							log.Printf("Recovered from panic in HealthCheckWorker: %v", r)
						}
					}()
					sendHealthData(endpoint)
				}()
			}
		}
	}()
}

// TriggerImmediateHealthCheck sends an immediate health check to the server (non-blocking)
func TriggerImmediateHealthCheck() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic in TriggerImmediateHealthCheck: %v", r)
			}
		}()
		sendHealthData(endpoint)
	}()
}
