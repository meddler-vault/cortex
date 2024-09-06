package bootstrap

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"

	httpSSL "net/http"

	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/meddler-vault/cortex/logger"

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

// DeleteMinIOFolder checks if a folder (prefix) exists in MinIO and deletes it fully if it does.
func DeleteMinIOFolder(ctx context.Context, client *minio.Client, bucketName, folderPath string) error {

	err := client.RemoveObject(ctx, bucketName, folderPath, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func skipMinioSSL() httpSSL.RoundTripper {
	// Keep TLS config.
	tlsConfig := &tls.Config{}

	tlsConfig.InsecureSkipVerify = true
	var transport httpSSL.RoundTripper = &httpSSL.Transport{
		Proxy: httpSSL.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsConfig,
		// Set this value so that the underlying transport round-tripper
		// doesn't try to auto decode the body of objects with
		// content-encoding set to `gzip`.
		//
		// Refer:
		//    https://golang.org/src/net/http/transport.go?h=roundTrip#L1843
		DisableCompression: true,
	}
	return transport

}

// / SyncDirToStorage syncs the given directory to the specified MinIO bucket and folder.
func SyncDirToStorage(bucketName string, folder string, dirPath string, stopAfterError bool, replace bool) (err error) {

	dirPath, err = filepath.Abs(dirPath)
	if err != nil {
		return err
	}
	dirPath += "/"

	ctx := context.Background()
	endpoint := populateStringFromEnv("MINIOURL", "localhost:9000")
	accessKeyID := populateStringFromEnv("MINIO_ACCESSKEY", "MEDDLER")
	secretAccessKey := populateStringFromEnv("MINIO_SECRET", "SUPERDUPERSECRET")
	useSSL := populateBoolFromEnv("MINIO_SECURE", true)
	region := populateStringFromEnv("MINIO_REGION", "india")

	logger := log.New(os.Stdout, "INFO: ", log.LstdFlags)
	logger.Println("Minio", endpoint, accessKeyID, secretAccessKey, useSSL, region)

	// Initialize MinIO client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Region:    region,
		Creds:     credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:    useSSL,
		Transport: skipMinioSSL(),
	})
	if err != nil {
		return err
	}

	policy, err := minioClient.GetBucketPolicy(context.Background(), bucketName)
	if err != nil {

		return err

	}

	// Print the retrieved policy
	fmt.Println("Bucket Policy:", policy, err)

	exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
	if errBucketExists != nil {
		return errBucketExists
	}

	if exists {

	} else {
		if err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	// Delete whole folder
	if replace {
		err = DeleteMinIOFolder(ctx, minioClient, bucketName, folder)
		log.Println("Deletion of folder", folder, err)
	}

	uploadFunc := func(path string, info os.FileInfo) error {
		// Calculate relative path from dirPath to the current file
		relPath, err := filepath.Rel(dirPath, path)
		// if err != nil {
		// 	return err
		// }

		// Construct object path within the bucket
		objectName := filepath.Join(folder, relPath)

		// Generate the relative path for the object name

		objectName = strings.TrimPrefix(objectName, string(filepath.Separator))

		// Upload the file
		logger.Println("Uploading", bucketName, objectName, path)
		_, err = minioClient.FPutObject(ctx, bucketName, objectName, path, minio.PutObjectOptions{})
		if err != nil {
			log.Println("Error", err)
			return err
		}

		return nil
	}

	onWalkFunc := func(path string, info os.FileInfo, err error) error {

		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				// return filepath.SkipDir
			}
			// return nil
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if uploadErr := uploadFunc(path, info); uploadErr != nil {
				if stopAfterError {
					return uploadErr
				}
			}
		}
		return nil
	}

	err = filepath.Walk(dirPath, onWalkFunc)
	return err
}

// SyncStorageToDir ()
func ExportDirToStorage(
	host string,
	accesskey string,
	secretkey string,
	secureConnection bool,
	region string,
	volumeMountPath string,
	bucketName string, dirPath string, stopAfterError bool, replace bool) (err error) {

	// Initialize minio client object.

	logger.Println("minio-config ExportDirToStorage", bucketName, host, accesskey, secretkey, secureConnection, region, dirPath)
	minioClient, err := minio.New(host, &minio.Options{
		Region:    region,
		Creds:     credentials.NewStaticV4(accesskey, secretkey, ""),
		Secure:    secureConnection,
		Transport: skipMinioSSL(),
	})

	logger.Println("minioClient", err)
	if err != nil {

		log.Println("Error", err)
		return
	}
	//

	ctx := context.Background()
	logger.Println("minioClient", "ctx", ctx)

	exists, err := minioClient.BucketExists(ctx, bucketName)

	if err != nil {
		logger.Println("Error", err)
		return
	}
	if !exists {
		return errors.New("invalid bucket name")
	}

	uploadFunc := func(path string, info os.FileInfo) error {
		// Calculate relative path from dirPath to the current file

		relPath, err := filepath.Rel(volumeMountPath, path)
		if err != nil {
			return err
		}

		// Construct object path within the bucket
		objectName := filepath.Join(dirPath, relPath)

		// Generate the relative path for the object name

		objectName = strings.TrimPrefix(objectName, string(filepath.Separator))

		// Upload the file
		logger.Println("Uploading", bucketName, objectName, path)
		_, err = minioClient.FPutObject(ctx, bucketName, objectName, path, minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		})
		if err != nil {
			log.Println("Error", err)
			return err
		}

		return nil
	}

	onWalkFunc := func(path string, info os.FileInfo, err error) error {
		logger.Println("onWalkFunc", "path", path, "info", info)

		if info == nil {
			return errors.New("No such path found to export")
		}
		// Skip hidden files and directories
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				// return filepath.SkipDir
			}
			// return nil
		}
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if uploadErr := uploadFunc(path, info); uploadErr != nil {
				if stopAfterError {
					return uploadErr
				}
			}
		}
		return nil
	}

	log.Println("walking path", volumeMountPath)
	err = filepath.Walk(volumeMountPath, onWalkFunc)
	return err

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

	useSSL := populateBoolFromEnv("MINIO_SECURE", true)

	// Initialize minio client object.
	region := populateStringFromEnv("MINIO_REGION", "india")
	// Initialize minio client object.

	logger.Println("Minio", endpoint, accessKeyID, secretAccessKey, useSSL, region)
	minioClient, err := minio.New(endpoint, &minio.Options{
		Region:    region,
		Creds:     credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure:    useSSL,
		Transport: skipMinioSSL(),
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

func isFile(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		return false
	}

	return !info.IsDir()
}
func isDir(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		return false
	}

	return info.IsDir()
}
func relativePath(p1, p2 string) (string, bool, error) {
	// Get the absolute paths of p1 and p2
	absP1, err := filepath.Abs(p1)
	if err != nil {
		return "", false, err
	}

	absP2, err := filepath.Abs(p2)
	if err != nil {
		return "", false, err
	}

	// Check if the absolute paths are the same
	if absP1 == absP2 {
		// Return the base name of the path (the filename)
		return filepath.Base(p1), true, nil
	}
	// Get the relative path of p1 with respect to p2
	relPath, err := filepath.Rel(absP2, absP1)
	if err != nil {
		return "", false, err
	}

	return relPath, false, nil
}

// SyncStorageToDir ()
func SyncMountVolumedToHost(
	host string,
	accesskey string,
	secretkey string,
	secureConnection bool,
	region string,
	volumeMountPath string,
	bucketName string, dirPath string, objectPath string, stopAfterError bool, replace bool) (folderMountPoint string, fileMountPoint string, err error) {

	// if dirPath == "" {
	// 	if objectPath == "" {
	// 		return "", "", errors.New("invalid file for volume mount")
	// 	} else {
	// 		dirPath = objectPath
	// 	}
	// } else {
	// 	if objectPath == "" {
	// 		// Do nothing
	// 	} else {
	// 		dirPath = filepath.Join(dirPath, objectPath)
	// 	}
	// }

	dirPath = filepath.Join(dirPath, objectPath)
	logger.Println("DirPath:", dirPath)
	// dirPath = strings.Trim(dirPath, " ")
	// objectPath = strings.Trim(objectPath, " ")

	logger.Println("SyncStorageToDir", volumeMountPath, dirPath, objectPath)

	if err != nil {
		return
	}

	ctx := context.Background()

	// Initialize minio client object.

	// Initialize minio client object.

	logger.Println("minio-config", bucketName, host, accesskey, secretkey, secureConnection, region, dirPath, objectPath)
	minioClient, err := minio.New(host, &minio.Options{
		Region:    region,
		Creds:     credentials.NewStaticV4(accesskey, secretkey, ""),
		Secure:    secureConnection,
		Transport: skipMinioSSL(),
	})

	// minioClient.TraceOn(os.Stdout)

	if err != nil {

		return
	}

	//
	exists, err := minioClient.BucketExists(ctx, bucketName)

	if err != nil {
		return
	}

	if exists {

		listObjectsChann := minioClient.ListObjects(context.Background(), bucketName, minio.ListObjectsOptions{
			Recursive: true,
			Prefix:    dirPath,
		})

		for obj := range listObjectsChann {

			var relPath string
			var isFile bool
			relPath, isFile, err = relativePath(obj.Key, dirPath)
			if err != nil {
				return
			}

			filePath := filepath.Join(volumeMountPath, relPath)

			logger.Println("Listing-path", dirPath, relPath, filePath)
			logger.Println("Listing", volumeMountPath, bucketName, obj.Key, filePath)
			err = minioClient.FGetObject(context.Background(), bucketName, obj.Key, filePath, minio.GetObjectOptions{})
			if err != nil && stopAfterError {
				return
			}

			if isFile {
				return volumeMountPath, filePath, nil
			}

		}

		return volumeMountPath, "", nil

		// errorCh := minioClient.FGetObject(context.Background(), bucketName, objectsCh, minio.RemoveObjectsOptions{})

	} else {
		err = errors.New("bucket does not exists")
		return
	}

}

// Git Functionality
func cloneRepositorySSH(url string, path string, privatekey string, password string) (repository *git.Repository, err error) {

	print("cloneRepositorySSH", privatekey, password)
	publicKeys, err := ssh.NewPublicKeys("git", []byte(privatekey), password)

	repository, err = git.PlainClone(path, false, &git.CloneOptions{
		// The intended use of a GitHub personal access token is in replace of your password
		// because access tokens can easily be revoked.
		// https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/
		Auth:     publicKeys,
		URL:      url,
		Progress: os.Stdout,
	})

	return

}

func cloneRepositoryAuth(url string, path string, username string, password string) (repository *git.Repository, err error) {

	println(path, username, password)
	repository, err = git.PlainClone(path, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: password,
		},
		URL:      url,
		Progress: os.Stdout,
	})

	return

}

func cloneRepositoryToken(url string, path string, username string, token string) (repository *git.Repository, err error) {

	if username == "" {
		username = "dummy_username"
	}
	repository, err = git.PlainClone(path, false, &git.CloneOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: token,
		},
		URL:      url,
		Progress: os.Stdout,
	})

	return

}

func cloneRepository(url string, path string, ref string, auth *transport.AuthMethod, gitdepth int) (repository *git.Repository, err error) {

	log.Println("auth_data", *auth)
	repository, err = git.PlainClone(path, false, &git.CloneOptions{
		URL:           url,
		Progress:      os.Stdout,
		Depth:         gitdepth,
		SingleBranch:  (len(ref)) > 0,
		ReferenceName: plumbing.ReferenceName(ref), // Assuming startCommitID is a tag
		Auth:          *auth,
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
func Clone(url string, path string, auth_mode string, username string, password string, gitref string, gitdepth int) (repository *git.Repository, err error) {

	log.Println("Clone()",
		url,
		path,
		auth_mode,
		username,
		password,
	)
	RemoveContents(path)

	var auth transport.AuthMethod = nil

	if auth_mode == NOAUTH {

	} else if auth_mode == BASICAUTH {
		log.Println("auth_mode", auth_mode)

		auth = &http.BasicAuth{
			Username: username,
			Password: password,
		}
		// repository, err = cloneRepositoryAuth(url, path, username, password)

	} else if auth_mode == TOKEN {
		auth = &http.TokenAuth{
			Token: password,
		}
		// repository, err = cloneRepositoryToken(url, path, username, password)

	} else if auth_mode == PRIVATEKEY {
		auth, err = ssh.NewPublicKeys("git", []byte(password), password)

		if err != nil {
			RemoveContents(path)

			return
		}
		// repository, err = cloneRepositorySSH(url, path, username, password)

	} else {
		// repository, err = cloneRepository(url, path)
	}

	repository, err = cloneRepository(url, path, gitref, &auth, gitdepth)

	// if err != nil, perform further operations

	if err != nil {
		RemoveContents(path)

		return
	}
	return
}
