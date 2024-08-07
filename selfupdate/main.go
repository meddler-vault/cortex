package selfupdate

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"

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

func Update() (string, string, error) {
	repo := "meddler-vault/cortex" // Replace with your repository
	platform := runtime.GOOS
	arch := runtime.GOARCH

	release, err := getLatestReleaseDetails(repo)
	if err != nil {
		log.Fatalf("Error fetching latest release: %v", err)
		return "", "", err

	}

	downloadURL, err := findDownloadURL(release, platform, arch)
	if err != nil {
		log.Fatalf("Error finding download URL: %v", err)
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
	}
	return err
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
