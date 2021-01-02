package bootstrap

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func init() {
}

// PopulateStr Populates from flag + env
func PopulateStr(key string, defaultValue string, description string) *string {
	return flag.String(key, populateStringFromEnv(key, defaultValue), description)
}

// PopulateBool Populates from flag + env
func PopulateBool(key string, defaultValue bool, description string) *bool {
	return flag.Bool(key, populateBoolFromEnv(key, defaultValue), description)
}

// PopulateInt Populates from flag + env
func PopulateInt(key string, defaultValue int, description string) *int {
	return flag.Int(key, populateIntFromEnv(key, defaultValue), description)

}

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
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func populateIntFromEnv(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

//RemoveContents ()
func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

//SyncDirToStorage ()
func SyncDirToStorage(bucketName string, dirPath string, stopAfterError bool) (err error) {

	ctx := context.Background()
	endpoint := "localhost:9000"
	accessKeyID := "MEDDLER"
	secretAccessKey := "SUPERDUPERSECRET"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return
	}

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
		_, err := minioClient.FPutObject(ctx, bucketName, path, path, minio.PutObjectOptions{})
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
