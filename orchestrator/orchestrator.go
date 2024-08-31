package main

import (
	"github.com/meddler-vault/cortex/db"
)

func init() {

}

func main() {

	// processEnqueuedJobs()

}

func processEnqueuedJobs() {
	// db.UpdateTaskResult(())
	db.UpdateTaskStatusInDraft("66d332e0badadf06c88b70fd")
}
