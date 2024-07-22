package main

import (
	"log"

	bootstrap "github.com/meddler-vault/cortex/bootstrap"
)

func main() {

	result := bootstrap.SSyncDirToStorage("minio-vapt", "0301bb5c-9a2b-4183-b4a6-47e5eb1f4b20/status", "/tmp/poop", false, true)
	log.Println(result)
}
