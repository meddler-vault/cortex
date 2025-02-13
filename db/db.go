package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
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

func getMongoCollection(collectionName string) (error, *mongo.Collection) {
	client, err := getMongoClient()
	if err != nil {
		return err, nil
	}

	MONGO_DB := getenvStr("MONGO_DB", "secoflex")

	collection := client.Database(MONGO_DB).Collection(collectionName)

	return nil, collection

}

func UpdateJobResult(taskResult bootstrap.TaskResult) error {
	var err error

	updateRes := bson.M{
		"exec_status":      taskResult.TaskStatus,
		"message":          taskResult.Message,
		"watchdog_version": taskResult.WatchdogVersion,
		"worker_id":        taskResult.WorkerId,
		"meta_data":        taskResult.MetaData,
	}

	err = updateBuildJobStatus(taskResult.Identifier, updateRes)

	return err

}

// Update task result
func UpdateTaskResult(taskResult bootstrap.TaskResult) error {
	var err error

	updateRes := bson.M{
		"exec_status":      taskResult.TaskStatus,
		"message":          taskResult.Message,
		"watchdog_version": taskResult.WatchdogVersion,

		"deployed":  false,
		"worker_id": taskResult.WorkerId,
		"meta_data": taskResult.MetaData,
	}

	err = updateTaskStatusInDraft(taskResult.Identifier, updateRes)

	return err

}

func UpdateTaskResultOld(subject string, taskResult bootstrap.TaskResult) error {

	prefixSplit := strings.Split(subject, bootstrap.RESULT_MESSAGE_QUEUE_SUBJECT_PREFIX)

	var collectionName string

	if len(prefixSplit) > 1 {
		collectionName = prefixSplit[1]
	}

	client, err := getMongoClient()
	if err != nil {
		return err
	}

	MONGO_DB := getenvStr("MONGO_DB", "secoflex")

	_id, err := primitive.ObjectIDFromHex(taskResult.Identifier)
	if err != nil {
		return err
	}

	// Hard coding the collection name for now
	collectionName = "job_drafts"

	collection := client.Database(MONGO_DB).Collection(collectionName)

	responseStruct := bson.M{
		"exec_status":      taskResult.TaskStatus,
		"message":          taskResult.Message,
		"watchdog_version": taskResult.WatchdogVersion,
		"worker_id":        taskResult.WorkerId,
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
		var updatedDoc bson.M
		// result, err = collection.UpdateByID(ctx, _id, bson.M{"$set": responseStruct})

		findOpts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)

		err := collection.FindOneAndUpdate(
			ctx,
			bson.M{"_id": _id},
			bson.M{"$set": responseStruct},
			findOpts,
		).Decode(&updatedDoc)

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

// ReplaceVariables replaces variables in the format {variable_name} with their corresponding values from the vars map.
func ReplaceVariables(s string, vars map[string]string) string {
	re := regexp.MustCompile(`{(\w+)}`) // Regular expression to find {variable_name}
	return re.ReplaceAllStringFunc(s, func(match string) string {
		key := match[1 : len(match)-1] // Extract variable name without braces
		if value, ok := vars[key]; ok {
			return value
		}
		return match // If variable not found, return the original match
	})
}

func updateBuildJobStatus(id string, updateRes bson.M) error {
	err, collection := getMongoCollection("builds_executor")
	if err != nil {
		return err
	}

	_id, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	doc := collection.FindOneAndUpdate(context.Background(), bson.M{

		"_id": _id,
	},
		bson.M{"$set": updateRes},
	)

	if doc == nil {
		return err
	}

	log.Println("update-result", _id)

	return nil
}

func updateTaskStatusInDraft(id string, updateRes bson.M) error {
	err, collection := getMongoCollection("job_drafts")

	if err != nil {
		return err
	}

	_id, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return err
	}

	doc := collection.FindOneAndUpdate(context.Background(), bson.M{

		"_id": _id,
	},
		bson.M{"$set": updateRes},
	)

	if doc == nil {
		return err
	}

	log.Println("update-result", _id)

	return nil

}

// Query and process documents
func ProcessDocuments() {

	// handle erros , disconnection ,etc as well
	_, collection := getMongoCollection("jobs")

	_id, _ := primitive.ObjectIDFromHex("66cce94a6eaae4d1e23e7db0")
	query := getQueryToFetchForwardAdjacentStepForJobId(_id)
	cursor, err := collection.Aggregate(context.Background(), query)

	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.Background())

	// Iterate over the cursor and print the results
	for cursor.Next(context.Background()) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
	}
	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	for {

		break

		// Find and atomically update one document to mark it as processing
		filter := bson.M{"exec_status": bootstrap.ENQUEUED}
		// update := bson.M{"$set": bson.M{"status": "processing", "updatedAt": time.Now()}}
		update := bson.M{"$set": bson.M{"exec_status": bootstrap.ENQUEUED}}

		// findOpts := options.FindOneAndUpdate().SetReturnDocument(options.After)
		findOpts := options.FindOneAndUpdate().SetSort(bson.M{"_id": -1}).SetReturnDocument(options.After)

		var doc bson.M
		err := collection.FindOneAndUpdate(context.TODO(), filter, update, findOpts).Decode(&doc)

		if err == mongo.ErrNoDocuments {
			fmt.Println("No new documents found, sleeping...")
			time.Sleep(5 * time.Second)
			continue
		} else if err != nil {
			log.Println("Error finding document:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		replacementID, ok := doc["_id"].(primitive.ObjectID)
		if !ok {
			log.Println("Error: 'request' field is not of type bson.M")
			return
		}
		// Example variables map
		vars := map[string]string{
			"id": replacementID.Hex(),
		}
		// fmt.Println("Processing document:", doc["_id"])
		// fmt.Println("Processing document:", doc["request"])

		request, ok := doc["request"].(bson.M)
		if !ok {
			log.Println("Error: 'request' field is not of type bson.M")
			return
		}

		job_config, ok := request["job_config"].(bson.M)
		if !ok {
			log.Println("Error: 'job_config' field is not of type bson.M")
			return
		}

		// Adjust type assertion for dependencies
		dependenciesRaw, ok := job_config["dependencies"].(bson.A)
		if !ok {
			log.Println("Error: 'dependencies' field is not of type []interface{}")
			fmt.Println("dependencies:", job_config["dependencies"])
			return
		}
		// If to replace it with..will be the current object's id

		// Iterate over each dependency
		// Convert each item to map[string]string
		// var dependencies []map[string]string
		for index, item := range dependenciesRaw {
			if depMap, ok := item.(bson.M); ok {
				depMapStr := make(map[string]string)
				for key, value := range depMap {
					if strVal, ok := value.(string); ok {
						depMapStr[key] = strVal

						value = ReplaceVariables(strVal, vars)
					} else {
						// log.Printf("Warning: value for key %s is not a string, value: %v", key, value)
					}

					depMap[key] = value
					log.Println(key, ":", value)
				}
				dependenciesRaw[index] = depMap
				// dependencies = append(dependencies, depMapStr)
			} else {
				log.Printf("Warning: item %v is not of type bson.M", item)
			}
		}

		fmt.Println("dependencies:", dependenciesRaw)
		// fmt.Println("Processing document:", doc)

		time.Sleep(5 * time.Second)

		// Simulate processing the document
		// if err := publishToNATS(natsConn, doc); err != nil {
		// 	log.Println("Error publishing to NATS:", err)
		// 	continue
		// }

		// Mark document as processed
		// _, err = collection.UpdateOne(context.TODO(), bson.M{"_id": doc["_id"]}, bson.M{"$set": bson.M{"status": "processed"}})
		// if err != nil {
		// 	log.Println("Error marking document as processed:", err)
		// } else {
		// 	fmt.Println("Document marked as processed:", doc["_id"])
		// }
	}
}

func getQueryToFetchForwardAdjacentStepForJobId(jobId primitive.ObjectID) []bson.D {
	// Define the MongoDB aggregation pipeline as a slice of bson.D (documents)
	return []bson.D{
		{
			bson.E{Key: "$match", Value: bson.D{{Key: "_id", Value: jobId}}},
		},
		{
			bson.E{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "tasks"},
				{Key: "pipeline", Value: bson.A{}}, // No additional match or conditions
				{Key: "as", Value: "tasks"},
			}},
		},
		{
			bson.E{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$tasks"},
				{Key: "preserveNullAndEmptyArrays", Value: false},
			}},
		},
		{
			bson.E{Key: "$replaceRoot", Value: bson.D{
				{Key: "newRoot", Value: bson.D{
					{Key: "$mergeObjects", Value: bson.A{
						"$tasks",
						bson.D{{Key: "refrence_task_id", Value: "$refrence_task_id"}},
					}},
				}},
			}},
		},
		{
			bson.E{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "tasks"},
				{Key: "localField", Value: "config.ingested_results.task_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "ingested_results"},
			}},
		},
		{
			bson.E{Key: "$set", Value: bson.D{
				{Key: "identifier", Value: "$config.identifier"},
				{Key: "from_identifier", Value: "$ingested_results.config.identifier"},
				{Key: "from_id", Value: "$ingested_results._id"},
			}},
		},
		{
			bson.E{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$ingested_results"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}},
		},
		{
			bson.E{Key: "$project", Value: bson.D{
				{Key: "from_id", Value: "$ingested_results._id"},
				{Key: "to_id", Value: "$_id"},
				{Key: "from_identifier", Value: "$ingested_results.config.identifier"},
				{Key: "to_identifier", Value: "$identifier"},
				{Key: "refrence_task_id", Value: "$refrence_task_id"},
			}},
		},
		{
			bson.E{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$from_id"},
				{Key: "refrence_task_id", Value: bson.D{{Key: "$first", Value: "$refrence_task_id"}}},
				{Key: "from_identifier", Value: bson.D{{Key: "$first", Value: "$from_identifier"}}},
				{Key: "edges_ids", Value: bson.D{{Key: "$push", Value: "$to_id"}}},
				{Key: "edges_identifiers", Value: bson.D{{Key: "$push", Value: "$to_identifier"}}},
			}},
		},
		{
			bson.E{Key: "$match", Value: bson.D{
				{Key: "_id", Value: bson.D{{Key: "$ne", Value: nil}}},
			}},
		},
		{
			bson.E{Key: "$match", Value: bson.D{
				{Key: "$expr", Value: bson.D{
					{Key: "$eq", Value: bson.A{"$_id", "$refrence_task_id"}},
				}},
			}},
		},
	}
}
