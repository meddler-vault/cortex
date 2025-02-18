package selfupdate

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/meddler-vault/cortex/bootstrap"
	"github.com/minio/selfupdate"
)

// Define a struct to parse the GitHub API response for releases
type Release struct {
	TagName string `json:"tag_name"`

	Assets []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// Function to get the latest release details from GitHub
func getLatestReleaseDetails(repo string) (Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("failed to fetch release: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return Release{}, err
	}

	return release, nil
}

// Function to find the download URL for the specified platform and architecture
func findDownloadURL(release Release, platform, arch string) (string, error) {
	for _, asset := range release.Assets {
		fileName := platform + "-" + arch
		if asset.Name == fileName {

			return asset.DownloadURL, nil
		}
	}
	return "", fmt.Errorf("no matching binary found for platform: %s and arch: %s", platform, arch)
}

// Do not change this logic
func DoUpdateInBetweenRuntimeCheck(currentVersion string) error {

	if bootstrap.DEBUG {
		return errors.New("debug mode enabled...skipping update and restaet!")
	}
	log.Println("++doUpdateInBetweenRuntimeCheck")
	_, version, err := Update(currentVersion)
	if err != nil {
		// Handle error
		log.Println("+++++++ [[No Force Restarting Startup]] +++++++", err)
		return err
	} else {
		log.Println("+++++++ [[Force Restarting Startup]] +++++++", currentVersion, " -->", version)
		return nil

	}

}

func Update(currentVersion string) (string, string, error) {
	repo := "meddler-vault/cortex" // Replace with your repository
	platform := runtime.GOOS
	arch := runtime.GOARCH

	release, err := getLatestReleaseDetails(repo)
	if err != nil {
		log.Println("Error fetching latest release: %v", err)
		return "", "", err

	}

	if release.TagName == currentVersion {
		return "", "", errors.New("no update required ")
	}

	downloadURL, err := findDownloadURL(release, platform, arch)
	if err != nil {
		log.Println("Error finding download URL: %v", err)
		return "", "", err
	}

	fmt.Printf("Download URL for %s-%s binary: %s version: %s \n\n ", platform, arch, downloadURL, release.TagName)

	err = doUpdate(downloadURL)

	return downloadURL, release.TagName, err
}

func doUpdate(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		// error handling
		log.Println("Error upgrade", err)
		return err

	}
	log.Println("Success upgrade", err)

	return nil
}

func RestartApp() error {
	cmd := exec.Command(os.Args[0], "--restarted")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Start the new process
	err := cmd.Start()
	if err != nil {
		return err
	}

	// Wait for the new process to be started
	err = cmd.Wait()
	if err != nil {
		return err
	}

	// Exit the current process
	os.Exit(0)

	return nil
}

func rForceQuit() {

	// logger.Println("")
	log.Println("+++++++ Force Restarting app +++++++")
	os.Exit(0)

}
func ForceQuit() {
	os.Exit(1)

	// if len(os.Args) > 1 && os.Args[1] == "new" {
	// 	// This is the new binary; do your work here
	// 	fmt.Println("New binary running.")
	// 	select {} // Keep running indefinitely
	// }

}
