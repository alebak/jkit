package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// latestReleaseURL is the GitHub API endpoint for the latest release.
const latestReleaseURL = "https://api.github.com/repos/alebak/jkit/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update jkit to the latest version",
	Long:  `Checks GitHub Releases for a newer version and self-updates the binary.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get latest release info
		release, err := fetchLatestRelease()
		if err != nil {
			return fmt.Errorf("checking for updates: %w", err)
		}

		latestVersion := strings.TrimPrefix(release.TagName, "v")
		if latestVersion == version || version == "dev" {
			cmd.Printf("Already up to date (%s)\n", release.TagName)
			return nil
		}

		// Find the right asset for this platform
		assetName := fmt.Sprintf("jkit-%s-%s", runtime.GOOS, runtime.GOARCH)
		var downloadURL string
		for _, a := range release.Assets {
			if a.Name == assetName {
				downloadURL = a.BrowserDownloadURL
				break
			}
		}
		if downloadURL == "" {
			return fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, release.TagName)
		}

		// Get current binary path
		binPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("finding current binary: %w", err)
		}
		binPath, err = filepath.EvalSymlinks(binPath)
		if err != nil {
			return fmt.Errorf("resolving binary path: %w", err)
		}

		// Download new binary
		cmd.Printf("Downloading %s...\n", release.TagName)
		resp, err := http.Get(downloadURL)
		if err != nil {
			return fmt.Errorf("downloading update: %w", err)
		}
		defer resp.Body.Close()

		tmpPath := binPath + ".new"
		tmp, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
		}

		if _, err := io.Copy(tmp, resp.Body); err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("writing update: %w", err)
		}
		tmp.Close()

		// Replace current binary
		if err := os.Rename(tmpPath, binPath); err != nil {
			os.Remove(tmpPath)
			return fmt.Errorf("installing update: %w", err)
		}

		cmd.Printf("Updated to %s\n", release.TagName)
		return nil
	},
}

func fetchLatestRelease() (*githubRelease, error) {
	resp, err := http.Get(latestReleaseURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing release info: %w", err)
	}
	return &release, nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
