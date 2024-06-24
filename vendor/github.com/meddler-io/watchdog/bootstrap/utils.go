package bootstrap

import (
	"context"
	"flag"
	"log"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/meddler-io/watchdog/logger"

	"os"
	"path/filepath"
	"strconv"
	"strings"

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

// RemoveContents ()
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

// SyncDirToStorage ()
func SyncDirToStorage(bucketName string, dirPath string, stopAfterError bool, replace bool) (err error) {

	dirPath, err = filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	dirPath += "/"

	ctx := context.Background()
	endpoint := populateStringFromEnv("MINIOURL", "localhost:9000")
	accessKeyID := populateStringFromEnv("MINIO_ACCESSKEY", "MEDDLER")
	secretAccessKey := populateStringFromEnv("MINIO_SECRET", "SUPERDUPERSECRET")

	useSSL := populateBoolFromEnv("MINIO_SECURE", false)

	region := populateStringFromEnv("MINIO_REGION", "india")

	logger.Println("Minio", endpoint, accessKeyID, secretAccessKey, useSSL, region)

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
						logger.Fatalln(object.Err)
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
			logger.Fatalln(err)
			return
		}
	} else {
		logger.Println("Successfully created ", bucketName)
	}

	uploadFunc := func(path string, info os.FileInfo) error {

		// Upload the zip file with FPutObject
		objPath := strings.SplitN(path, dirPath, 2)[1]
		logger.Println("Uploading", path, objPath, dirPath)
		_, err := minioClient.FPutObject(ctx, bucketName, objPath, path, minio.PutObjectOptions{})
		// filename := filepath.Join(path, info.Name())
		// logger.Println("Uploading", info.Name(), err)
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

// SyncStorageToDir ()
func SyncStorageToDir(bucketName string, dirPath string, identifier string, stopAfterError bool, replace bool) (err error) {

	logger.Println("SyncStorageToDir")
	dirPath, err = filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	dirPath += "/"

	ctx := context.Background()
	endpoint := populateStringFromEnv("MINIOURL", "localhost:9000")

	accessKeyID := populateStringFromEnv("MINIO_ACCESSKEY", "MEDDLER")
	secretAccessKey := populateStringFromEnv("MINIO_SECRET", "SUPERDUPERSECRET")

	useSSL := populateBoolFromEnv("MINIO_SECURE", false)

	// Initialize minio client object.
	region := populateStringFromEnv("MINIO_REGION", "india")
	// Initialize minio client object.

	logger.Println("Minio", endpoint, accessKeyID, secretAccessKey, useSSL, region)
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

		listObjectsChann := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
			Recursive: true,
			Prefix:    "/",
		})

		for obj := range listObjectsChann {
			logger.Println(obj.Key)
			filePath := filepath.Join(dirPath, identifier, obj.Key)
			err = minioClient.FGetObject(context.Background(), bucketName, obj.Key, filePath, minio.GetObjectOptions{})
			if err != nil && stopAfterError {
				return err
			}
			logger.Println(err)

		}

		// errorCh := minioClient.FGetObject(context.Background(), bucketName, objectsCh, minio.RemoveObjectsOptions{})

		return
	}

	return

}

// Git Functionality
func cloneRepositorySSH(url string, path string, privatekey string, password string) (err error) {

	print("cloneRepositorySSH", privatekey, password)
	publicKeys, err := ssh.NewPublicKeys("git", []byte(privatekey), password)

	_, err = git.PlainClone(path, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth:     publicKeys,
		URL:      url,
		Progress: os.Stdout,
	})

	return

}

func cloneRepositoryAuth(url string, path string, username string, password string) (err error) {

	println(path, username, password)
	_, err = git.PlainClone(path, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
		URL:      url,
		Progress: os.Stdout,
	})

	return

}

func cloneRepositoryToken(url string, path string, username string, token string) (err error) {

	if username == "" {
		username = "dummy_username"
	}
	_, err = git.PlainClone(path, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: token,
		},
		URL:      url,
		Progress: os.Stdout,
	})

	return

}

func cloneRepository(url string, path string) (err error) {
	_, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})

	return
}

// url: HTTPS / SSH Url
//
// auth_mode: no_auth, ssh, token , password
//
// username: Auth. username / Private Key string
//
// password: Auth. password / Private Key password
func Clone(url string, path string, auth_mode string, username string, password string) (err error) {

	log.Println("Clone()",
		url,
		path,
		auth_mode,
		username,
		password,
	)
	RemoveContents(path)

	if auth_mode == NOAUTH {
		err = cloneRepository(url, path)

	} else if auth_mode == PASSWORD {
		err = cloneRepositoryAuth(url, path, username, password)

	} else if auth_mode == TOKEN {
		err = cloneRepositoryToken(url, path, username, password)

	} else if auth_mode == PRIVATEKEY {
		err = cloneRepositorySSH(url, path, username, password)

	} else {
		err = cloneRepository(url, path)
	}

	return err
}
