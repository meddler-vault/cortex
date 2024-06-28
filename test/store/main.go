package main

import (
	"log"

	bootstrap "github.com/meddler-vault/cortex/bootstrap"
)

func main() {

	result := bootstrap.SyncDirToStorage("minio-vapt", "test-folder-new", "/Users/apple/workspaces/meddler/cortex/", false, true)
	log.Println(result)
}
