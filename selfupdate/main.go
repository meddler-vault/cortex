package selfupdate

import (
	"net/http"

	"encoding/json"
	"fmt"
	"log"

	"strings"

	"github.com/minio/selfupdate"
)

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

// Define a struct to parse the GitHub API response
type Release struct {
	TagName string `json:"tag_name"`
}

// Function to get the latest release tag from GitHub
func getLatestReleaseTag(repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch release: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

// Function to check if the current version is outdated
func isVersionOutdated(currentVersion, latestVersion string) bool {
	// Remove "v" prefix for comparison if necessary
	currentVersion = strings.TrimPrefix(currentVersion, "v")
	latestVersion = strings.TrimPrefix(latestVersion, "v")

	// Compare versions. This is a simple comparison; use a library for semantic versioning if needed.
	return latestVersion > currentVersion
}

func Update() {
	repo := "meddler-vault/cortex" // Replace with your repository
	currentVersion := "1.0.0"      // Replace with your current version

	latestVersion, err := getLatestReleaseTag(repo)
	if err != nil {
		log.Fatalf("Error fetching latest release: %v", err)
	}

	fmt.Printf("Current Version: %s\n", currentVersion)
	fmt.Printf("Latest Version: %s\n", latestVersion)

	if isVersionOutdated(currentVersion, latestVersion) {
		fmt.Println("Your version is outdated.")
	} else {
		fmt.Println("Your version is up to date.")
	}
}
