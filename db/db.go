package db

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/meddler-vault/cortex/bootstrap"
	"github.com/meddler-vault/cortex/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance    *mongo.Client
	clientInstanceErr error
	mongoOnce         sync.Once
	maxRetryAttempts  = 3
	retryInitialDelay = 2 * time.Second
)

// Singleton pattern to create a single MongoDB client instance
func getMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		MONGO_URL := getenvStr("MONGO_URL", "mongodb://localhost")

		clientOptions := options.Client().ApplyURI(MONGO_URL)
		client, err := mongo.Connect(context.Background(), clientOptions)
		if err != nil {
			clientInstanceErr = err
			return
		}

		// Check the connection
		err = client.Ping(context.Background(), nil)
		if err != nil {
			clientInstanceErr = err
			return
		}

		clientInstance = client
	})

	return clientInstance, clientInstanceErr
}

func getenvStr(key string, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}

func UpdateTaskResult(collectionName string, taskResult bootstrap.TaskResult) error {
	client, err := getMongoClient()
	if err != nil {
		return err
	}

	MONGO_DB := getenvStr("MONGO_DB", "secoflex")

	_id, err := primitive.ObjectIDFromHex(taskResult.Identifier)
	if err != nil {
		return err
	}

	collection := client.Database(MONGO_DB).Collection(collectionName)

	responseStruct := bson.M{
		"exec_status":      taskResult.TaskStatus,
		"message":          taskResult.Message,
		"watchdog_version": taskResult.WatchdogVersion,
	}

	var response map[string]interface{}
	err = json.Unmarshal([]byte(taskResult.Response), &response)
	if err == nil {
		responseStruct["result"] = response
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var result *mongo.UpdateResult
	retryCount := 0
	for {
		result, err = collection.UpdateByID(ctx, _id, bson.M{"$set": responseStruct})
		if err == nil || retryCount >= maxRetryAttempts {
			break
		}
		retryCount++
		log.Println("Retrying update due to error:", err)
		time.Sleep(retryInitialDelay * time.Duration(retryCount))
	}

	if err != nil {
		return errors.New("failed to update task result after multiple attempts")
	}

	logger.Println("Task:", "Updated Results", taskResult.Identifier, result, err)

	logger.Println("collectionName", collectionName, result.MatchedCount, result.ModifiedCount)

	return nil
}
