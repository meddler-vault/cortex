package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/meddler-vault/cortex/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func populateStringFromEnv(key string, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func populateBoolFromEnv(key string, defaultVal bool) bool {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			logger.Fatalln("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func populateIntFromEnv(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			logger.Fatalln("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}
func main() {
	// err := syncDirToStorage("synctest", "/Users/meddler/Office/Workspaces/Secoflex/secoflex/modules/watchdog", false)
	err := syncDirToStorage("synctest", "./", false, true)
	// err := syncDirToStorage("synctest", "./", false, false)
	log.Println(err)
}

func cloneRepositoryAuth(url string, path string, username string, password string) {
	_, err := git.PlainClone("/tmp/foo", false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
		URL:      "https://github.com/go-git/go-git",
		Progress: os.Stdout,
	})

	if err == nil {
		return
	}

}
func cloneRepositoryToken(url string, path string, username string, token string) {
	_, err := git.PlainClone("/tmp/foo", false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: token,
		},
		URL:      "https://github.com/go-git/go-git",
		Progress: os.Stdout,
	})
	if err != nil {
		return
	}

}

func cloneRepository(url string, path string) {
	_, err := git.PlainClone("/tmp/foo", false, &git.CloneOptions{
		URL:      "https://github.com/go-git/go-git",
		Progress: os.Stdout,
	})
	if err != nil {
		return
	}
}

func syncDirToStorage(bucketName string, dirPath string, stopAfterError bool, replace bool) (err error) {

	dirPath, err = filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	dirPath += "/"

	ctx := context.Background()
	//
	endpoint := populateStringFromEnv("MINIOURL", "localhost:9000")
	accessKeyID := populateStringFromEnv("MINIO_ACCESSKEY", "MEDDLER")
	secretAccessKey := populateStringFromEnv("MINIO_SECRET", "SUPERDUPERSECRET")

	useSSL := populateBoolFromEnv("MINIO_SECURE", false)

	region := populateStringFromEnv("MINIO_REGION", "india")

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Region: region,
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
		// log.Printf("We already own %s\n", bucketName)
		if replace {

			objectsCh := make(chan minio.ObjectInfo)

			// Send object names that are needed to be removed to objectsCh
			// Send object names that are needed to be removed to objectsCh
			go func() {
				defer close(objectsCh)
				// List all objects from a bucket-name with a matching prefix.
				opts := minio.ListObjectsOptions{Prefix: "", Recursive: true}
				for object := range minioClient.ListObjects(context.Background(), bucketName, opts) {
					if object.Err != nil {
						log.Fatalln(object.Err)
					}
					objectsCh <- object
				}
			}()

			// Call RemoveObjects API
			errorCh := minioClient.RemoveObjects(context.Background(), bucketName, objectsCh, minio.RemoveObjectsOptions{})

			// Print errors received from RemoveObjects API
			for e := range errorCh {
				return e.Err
			}

			if err = minioClient.RemoveBucket(ctx, bucketName); err != nil {
				return err
			}

			if err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
				return err
			}
		}
	} else {
		if err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	//
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			// log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
			return
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	uploadFunc := func(path string, info os.FileInfo) error {

		// Upload the zip file with FPutObject
		objPath := strings.SplitN(path, dirPath, 2)[1]
		log.Println("Uploading", path, objPath, dirPath)
		_, err := minioClient.FPutObject(ctx, bucketName, objPath, path, minio.PutObjectOptions{})
		// filename := filepath.Join(path, info.Name())
		// log.Println("Uploading", info.Name(), err)
		if err != nil {
			return err
		}

		return nil

	}

	onWalkFunc := func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !info.IsDir() {
			// rel, _ := filepath.Rel(path, "a/c/t/file")
			if uploadErr := uploadFunc(path, info); uploadErr != nil {
				if stopAfterError {
					return uploadErr
				} else {
					return nil
				}
			}

		}
		return nil
	}

	err = filepath.Walk(dirPath, onWalkFunc)

	return

}
