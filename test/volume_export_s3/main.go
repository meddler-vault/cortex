package main

import (
	"log"

	bootstrap "github.com/meddler-vault/cortex/bootstrap"
)

func main() {

	result := bootstrap.ExportDirToStorage(
		"s3.meddler.io",
		"uaaGAF0jnXVHa7KV5eOa",
		"kiwty0-Xigruc-zyfnyj",
		true,
		"auto",
		"volume_export_test/hello",

		"minio-vapt", "/tmp/test", true, true)
	{
		log.Println(result)
	}
}
