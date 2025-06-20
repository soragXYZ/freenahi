// package github contains features for accessing repos on Github.
package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-version"
)

type Results struct {
	Version string `json:"name"`
}
type DockerImageVersion struct {
	Results []Results `json:"results"`
}

// AvailableUpdate return the version of the latest release and reports wether the update is newer.
func AvailableUpdate(gitHubOwner, githubRepo, localVersion string) (string, bool, error) {
	local, err := version.NewVersion(localVersion)
	if err != nil {
		return "", false, err
	}
	r, err := fetchGitHubLatest(gitHubOwner, githubRepo)
	if err != nil {
		return "", false, err
	}
	remote, err := version.NewVersion(r)
	if err != nil {
		return "", false, err
	}

	return remote.String(), local.LessThan(remote), nil
}

func fetchGitHubLatest(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	if r.StatusCode >= 400 {
		return "", fmt.Errorf("%s: %w", r.Status, errors.New("HTTP error"))
	}

	var info struct {
		TagName string `json:"tag_name"`
	}

	if err := json.Unmarshal(data, &info); err != nil {
		return "", err
	}
	return info.TagName, nil
}

func FetchDockerLatest(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/%s/tags/?page_size=1", owner, repo)
	r, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	if r.StatusCode >= 400 {
		return "", fmt.Errorf("%s: %w", r.Status, errors.New("HTTP error"))
	}

	var imageVersion DockerImageVersion
	if err := json.Unmarshal(data, &imageVersion); err != nil {
		return "", err
	}

	return imageVersion.Results[0].Version, nil
}
