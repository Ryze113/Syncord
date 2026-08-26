/*
 * SPDX-License-Identifier: GPL-3.0
 * Syncord Installer, a cross platform gui/cli app for installing Syncord
 * Copyright (c) 2023 Vendicated and SYNCORD contributors
 */

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	path "path/filepath"
	"strconv"
	"strings"
	"sync"
)

type GithubRelease struct {
	Name    string `json:"name"`
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetGithubRelease(url, fallbackUrl string) (*GithubRelease, error) {
	Log.Debug("Fetching", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Log.Error("Failed to create Request", err)
		return nil, err
	}

	req.Header.Set("User-Agent", UserAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Log.Error("Failed to send Request", err)
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode >= 300 {
		isRateLimitedOrBlocked := res.StatusCode == 401 || res.StatusCode == 403 || res.StatusCode == 429
		triedFallback := url == fallbackUrl

		if isRateLimitedOrBlocked && !triedFallback {
			Log.Error(fmt.Sprintf("Failed to fetch %s (status code %d). Trying fallback url %s", url, res.StatusCode, fallbackUrl))
			return GetGithubRelease(fallbackUrl, fallbackUrl)
		}

		err = errors.New(res.Status)
		Log.Error(url, "returned Non-OK status", err)
		return nil, err
	}

	var data GithubRelease

	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		Log.Error("Failed to decode GitHub JSON Response", err)
		return nil, err
	}

	return &data, nil
}

var RemoteVersion = "Unknown"
var LocalVersion = "Unknown"
var HasUpdate bool
var GithubError error
var GithubDoneChan chan bool
var IsDevInstall bool

func fetchRemoteVersion() (string, error) {
	Log.Debug("Fetching remote version from", VersionURL)

	req, err := http.NewRequest("GET", VersionURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgent)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		return "", errors.New(res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(body)), nil
}

func parseVersion(v string) []int {
	parts := strings.Split(v, ".")
	result := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil
		}
		result[i] = n
	}
	return result
}

func isVersionNewer(local, remote string) bool {
	// Simple string compare — if content differs, update
	return strings.TrimSpace(local) != strings.TrimSpace(remote)
}

func getLocalVersionPath() string {
	return path.Join(FilesDir, "..", "version.txt")
}

func getLocalVersion() string {
	data, err := os.ReadFile(getLocalVersionPath())
	if err != nil {
		return "0.0.0"
	}
	return strings.TrimSpace(string(data))
}

func saveLocalVersion(version string) {
	os.WriteFile(getLocalVersionPath(), []byte(version), 0644)
}

func InitGithubDownloader() {
	GithubDoneChan = make(chan bool, 1)

	IsDevInstall = os.Getenv("SYNCORD_DEV_INSTALL") == "1"
	Log.Debug("Is Dev Install: ", IsDevInstall)
	if IsDevInstall {
		GithubDoneChan <- true
		return
	}

	// Check local version from installed patcher
	f, err := os.Open(Patcher)
	if err == nil {
		scanner := bufio.NewScanner(f)
		if scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "// Syncord ") {
				LocalVersion = line[11:]
				Log.Debug("Existing hash is", LocalVersion)
			}
		}
		f.Close()
	}

	// If no hash found in patcher, use version.txt
	if LocalVersion == "Unknown" || LocalVersion == "" {
		LocalVersion = getLocalVersion()
	}

	go func() {
		defer func() {
			GithubDoneChan <- GithubError == nil
		}()

		remote, err := fetchRemoteVersion()
		if err != nil {
			GithubError = err
			Log.Error("Failed to fetch remote version:", err)
			return
		}

		RemoteVersion = remote
		Log.Debug("Remote version:", RemoteVersion, "Local:", LocalVersion)

		HasUpdate = isVersionNewer(LocalVersion, RemoteVersion)
		Log.Debug("Has update:", HasUpdate)
	}()
}

func downloadSyncordFiles() (retErr error) {
	Log.Debug("Downloading Syncord files...")

	pkgJsonFile := path.Join(FilesDir, "package.json")
	err := os.WriteFile(pkgJsonFile, []byte("{}"), 0644)
	if err != nil {
		Log.Warn("Failed to create", pkgJsonFile, err)
	}

	var wg sync.WaitGroup

	for _, fileName := range SyncordFiles {
		wg.Add(1)
		fileName := fileName
		go func() {
			defer wg.Done()
			url := SyncordDistURL + fileName
			Log.Debug("Downloading", url)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				Log.Error("Failed to create request for", fileName+":", err)
				retErr = err
				return
			}
			req.Header.Set("User-Agent", UserAgent)

			res, err := http.DefaultClient.Do(req)
			if err == nil && res.StatusCode >= 300 {
				err = errors.New(res.Status)
			}
			if err != nil {
				Log.Error("Failed to download", fileName+":", err)
				retErr = err
				return
			}
			defer res.Body.Close()

			outFile := path.Join(FilesDir, fileName)
			out, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				Log.Error("Failed to create", outFile+":", err)
				retErr = err
				return
			}
			defer out.Close()

			read, err := io.Copy(out, res.Body)
			if err != nil {
				Log.Error("Failed to write to", outFile+":", err)
				retErr = err
				return
			}
			contentLength := res.Header.Get("Content-Length")
			expected := strconv.FormatInt(read, 10)
			if expected != contentLength {
				err = errors.New(fmt.Sprintf("Unexpected end of input. Content-Length was %s, but I only read %s", contentLength, expected))
				Log.Error(err.Error())
				retErr = err
				return
			}
		}()
	}

	wg.Wait()
	Log.Debug("Done!")
	_ = FixOwnership(FilesDir)

	LocalVersion = RemoteVersion
	saveLocalVersion(RemoteVersion)
	return
}

func installLatestBuilds() (retErr error) {
	Log.Debug("Installing latest builds...")
	return downloadSyncordFiles()
}
