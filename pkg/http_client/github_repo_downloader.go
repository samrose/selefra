package http_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/selefra/selefra-provider-sdk/provider/schema"
	"os"
	"path/filepath"
	"time"
)

// ------------------------------------------------- --------------------------------------------------------------------

// GitHubRepoDownloaderOptions download github
type GitHubRepoDownloaderOptions struct {

	// Who owns the warehouse, org name or username
	Owner string

	// Name of warehouse
	Repo string

	// Which directory to download it to
	DownloadDirectory string

	// Whether to use a cache, using a cache can avoid repeat download in a short time
	// Open is suitable for large warehouse, when the warehouse itself is not large, you can download it directly
	CacheTime *time.Duration

	// It may take a while, but some messages will be sent to you if needed
	MessageChannel chan *schema.Diagnostics

	ProgressListener getter.ProgressTracker
}

type GitHubRepoDownloader struct {
}

func NewGitHubRepoDownloader() *GitHubRepoDownloader {
	return &GitHubRepoDownloader{}
}

func (x *GitHubRepoDownloader) Download(ctx context.Context, options *GitHubRepoDownloaderOptions) error {

	// check cache if use it
	if options.CacheTime != nil {
		if x.checkCache(options.DownloadDirectory, *options.CacheTime) {
			return nil
		}
	}

	targetUrl := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/main.zip", options.Owner, options.Repo)
	err := DownloadToDirectory(ctx, options.DownloadDirectory, targetUrl, options.ProgressListener)
	if err != nil {
		return err
	}

	if options.CacheTime != nil {
		if err := x.Save(options.DownloadDirectory, &GitHubRepoCacheMeta{DownloadTime: time.Now()}); err != nil {
			return err
		}
	}

	return nil
}

func (x *GitHubRepoDownloader) checkCache(downloadDirectory string, cacheTime time.Duration) bool {
	meta, err := x.ReadCacheMeta(downloadDirectory)
	if err != nil {
		return false
	}
	if time.Now().Sub(meta.DownloadTime) > cacheTime {
		return false
	}
	return true
}

// GitHubRepoCacheMeta Cache information from the github repository
type GitHubRepoCacheMeta struct {
	// The last download time of the repo
	DownloadTime time.Time `json:"download-time"`
}

// ReadCacheMeta the github repository cache
func (x *GitHubRepoDownloader) ReadCacheMeta(downloadDirectory string) (*GitHubRepoCacheMeta, error) {
	fileBytes, err := os.ReadFile(x.BuildCacheMetaFilePath(downloadDirectory))
	if err != nil {
		return nil, err
	}
	r := new(GitHubRepoCacheMeta)
	err = json.Unmarshal(fileBytes, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Save the github repository cache information
func (x *GitHubRepoDownloader) Save(downloadDirectory string, meta *GitHubRepoCacheMeta) error {

	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	metaFilePath := x.BuildCacheMetaFilePath(downloadDirectory)
	return os.WriteFile(metaFilePath, metaBytes, os.FileMode(0644))
}

// BuildCacheMetaFilePath The root path of the downloaded repository is followed by a cache metadata-related file
func (x *GitHubRepoDownloader) BuildCacheMetaFilePath(downloadDirectory string) string {
	return filepath.Join(downloadDirectory, ".selefra-cache-meta")
}

// ------------------------------------------------- --------------------------------------------------------------------
