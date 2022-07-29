/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// privateArtifactCmd represents the privateArtifact command
var privateArtifactCmd = &cobra.Command{
	Use:   "privateArtifact",
	Short: "Download artifacts from releases in private repositories",
	Long: `Download artifacts from releases in private repositories. For example:

TODO: Add example here.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get token from environment variable or else from flag
		token := os.Getenv("GITHUB_TOKEN")
		if token == "" {
			token = cmd.Flag("token").Value.String()
		}

		// Get the repo from the flag
		repo := cmd.Flag("repo").Value.String()

		// Get the version to download from the flag or else equal to "latest"
		version := cmd.Flag("version").Value.String()
		if version == "" {
			version = "latest"
		}

		// get the assetname
		assetName := cmd.Flag("assetName").Value.String()

		// Get asset id
		asset, err := getAsset(token, repo, assetName, version)
		//Download from github
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Asset id:", asset.ID)
		fmt.Println("Asset file name:", asset.Name)
		// Get the asset from github
		err = downloadAsset(asset, token, repo)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Downloaded asset:", asset.Name)
	},
}

func downloadAsset(asset GithubRepoAsset, token string, repo string) error {
	// Get the asset from github
	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s:@api.github.com/repos/%s/releases/assets/%d", token, repo, asset.ID), nil)
	if err != nil {
		return err
	}
	// add headers
	req.Header.Set("Accept", "application/octet-stream")
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	file, err := os.Create(asset.Name)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the body to file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func getAsset(token string, repo string, assetName string, version string) (GithubRepoAsset, error) {
	// Get the asset id
	req, err := NewGithubRequest("https://api.github.com/repos/"+repo+"/releases", token)
	if err != nil {
		return GithubRepoAsset{}, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GithubRepoAsset{}, err
	}
	defer resp.Body.Close()

	// Marshal as GithubRepoResponse
	var githubRepoResponses GithubRepoResponses
	err = json.NewDecoder(resp.Body).Decode(&githubRepoResponses)
	if err != nil {
		return GithubRepoAsset{}, err
	}

	var release GithubRepoResponse
	if version == "latest" {
		// Github should return the latest release first.
		release = githubRepoResponses[0]

	} else {
		// Find the release with the version
		for _, r := range githubRepoResponses {
			if r.TagName == version {
				release = r
			}
		}
	}

	if release.Assets == nil || len(release.Assets) == 0 {
		return GithubRepoAsset{}, fmt.Errorf("no assets found for latest release")
	}

	// Get the asset with the specified name
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return asset, nil
		}
	}
	return GithubRepoAsset{}, fmt.Errorf("no asset found with name: %s", assetName)
}

func NewGithubRequest(url string, token string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// add headers
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3.raw")

	return req, nil
}

func init() {
	rootCmd.AddCommand(privateArtifactCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// privateArtifactCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// privateArtifactCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	privateArtifactCmd.Flags().StringP("token", "t", "", "Github token")
	privateArtifactCmd.Flags().StringP("repo", "r", "", "Github repo")
	privateArtifactCmd.MarkFlagRequired("repo")
	privateArtifactCmd.Flags().StringP("assetName", "a", "", "Name of the asset")
	privateArtifactCmd.MarkFlagRequired("assetName")
	privateArtifactCmd.Flags().StringP("version", "v", "", "Version to download")
}

type GithubRepoResponses []GithubRepoResponse
type GithubRepoResponse struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeID          string            `json:"node_id"`
	TagName         string            `json:"tag_name"`
	TargetCommitish string            `json:"target_commitish"`
	Name            string            `json:"name"`
	Draft           bool              `json:"draft"`
	Prerelease      bool              `json:"prerelease"`
	CreatedAt       time.Time         `json:"created_at"`
	PublishedAt     time.Time         `json:"published_at"`
	Assets          []GithubRepoAsset `json:"assets"`
	TarballURL      string            `json:"tarball_url"`
	ZipballURL      string            `json:"zipball_url"`
	Body            string            `json:"body"`
}

type GithubRepoAsset struct {
	URL      string      `json:"url"`
	ID       int         `json:"id"`
	NodeID   string      `json:"node_id"`
	Name     string      `json:"name"`
	Label    interface{} `json:"label"`
	Uploader struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"uploader"`
	ContentType        string    `json:"content_type"`
	State              string    `json:"state"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}
