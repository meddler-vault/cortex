package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/meddler-io/watchdog/bootstrap"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	// err := syncDirToStorage("synctest", "/Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog", false)
	err := SyncStorageToDir("synctest", *bootstrap.INPUTDIR, "lol", false, true)
	// err := syncDirToStorage("synctest", "./", false, false)
	log.Println(err)
}

func populateStringFromEnv(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func SyncStorageToDir(bucketName string, dirPath string, identifier string, stopAfterError bool, replace bool) (err error) {

	dirPath, err = filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	dirPath += "/"

	ctx := context.Background()
	endpoint := populateStringFromEnv("MINIOURL", "localhost:9000")
	accessKeyID := populateStringFromEnv("MINIO_ACCESSKEY", "MEDDLER")
	secretAccessKey := populateStringFromEnv("MINIO_SECRET", "SUPERDUPERSECRET")

	useSSL = populateBoolFromEnv("MINIO_SECURE", false)

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return
	}
	//
	exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)

	if errBucketExists != nil {
		return errBucketExists
	}

	if exists {

		log.Println("Eglisting")

		listObjectsChann := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
			Recursive: true,
			Prefix:    "/",
		})

		for obj := range listObjectsChann {
			log.Println(obj.Key)
			filePath := filepath.Join(dirPath, identifier, obj.Key)
			err = minioClient.FGetObject(context.Background(), bucketName, obj.Key, filePath, minio.GetObjectOptions{})
			if err != nil && stopAfterError {
				return err
			}
			log.Println(err)

		}

		// errorCh := minioClient.FGetObject(context.Background(), bucketName, objectsCh, minio.RemoveObjectsOptions{})

		return
	}

	return

}
