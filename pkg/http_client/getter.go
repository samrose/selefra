package http_client

import (
	"context"
	"net/http"
	"os"
	"time"

	getter "github.com/hashicorp/go-getter"
)

type Detector struct {
	Name     string
	Detector getter.Detector
}

var (
	detectors = []getter.Detector{
		new(getter.GitHubDetector),
		new(getter.GitDetector),
		new(getter.S3Detector),
		new(getter.GCSDetector),
		new(getter.FileDetector),
	}

	decompressors = map[string]getter.Decompressor{
		"bz2": new(getter.Bzip2Decompressor),
		"gz":  new(getter.GzipDecompressor),
		"xz":  new(getter.XzDecompressor),
		"zip": new(getter.ZipDecompressor),

		"tar.bz2":  new(getter.TarBzip2Decompressor),
		"tar.tbz2": new(getter.TarBzip2Decompressor),

		"tar.gz": new(getter.TarGzipDecompressor),
		"tgz":    new(getter.TarGzipDecompressor),

		"tar.xz": new(getter.TarXzDecompressor),
		"txz":    new(getter.TarXzDecompressor),
	}

	getters = map[string]getter.Getter{
		"file":   new(getter.FileGetter),
		"gcs":    new(getter.GCSGetter),
		"github": new(getter.GitGetter),
		"git":    new(getter.GitGetter),
		"hg":     new(getter.HgGetter),
		"s3":     new(getter.S3Getter),
		"http":   httpGetter,
		"https":  httpGetter,
	}
)

var httpGetter = &getter.HttpGetter{
	ReadTimeout:           10 * time.Minute,
	MaxBytes:              1_000_000_000,
	XTerraformGetDisabled: true,
	//Client: &http.Client{
	//	CheckRedirect: func(req *http.Request, via []*http.Request) error {
	//		return nil
	//	},
	//},
	Header: http.Header{
		"User-Agent": []string{MyUserAgent()},
	},
	//DoNotCheckHeadFirst: true,
}

func DownloadToDirectory(ctx context.Context, saveDirectory, targetUrl string, progressListener getter.ProgressTracker, options ...getter.ClientOption) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	client := getter.Client{
		Src:           targetUrl,
		Dst:           saveDirectory,
		Pwd:           pwd,
		Mode:          getter.ClientModeDir,
		Detectors:     detectors,
		Decompressors: decompressors,
		Getters:       getters,
		Ctx:           ctx,
		// Extra options provided by caller to overwrite default behavior
		Options:          options,
		ProgressListener: progressListener,
	}

	return client.Get()
}

//func ModuleGet(ctx context.Context, installPath, url string, options ...getter.ClientOption) error {
//	pwd, _ := os.Getwd()
//	client := getter.Client{
//		Src:           url,
//		Dst:           installPath,
//		Pwd:           pwd,
//		Mode:          getter.ClientModeDir,
//		Detectors:     detectors,
//		Decompressors: decompressors,
//		Getters:       getters,
//		Ctx:           ctx,
//		// Extra options provided by caller to overwrite default behavior
//		Options: options,
//	}
//
//	if err := client.DownloadToDirectory(); err != nil {
//		return err
//	}
//	return nil
//}
