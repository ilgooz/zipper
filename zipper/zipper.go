// Package zipper responsible to download files over HTTP, compress them on
// the fly and pipe the zip output to given destination.
package zipper

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"time"
)

// DownloaderClient used to download files over HTTP.
var DownloaderClient = &http.Client{
	Timeout: 30 * time.Second,
}

// File represents a file that needs to be downloaded over HTTP.
type File struct {
	URL  string `json:"url"`
	Name string `json:"filename"`
}

// Download downloads files one by one, zips them on the fly and writes zip output to w as a stream.
// This is a pipe, writes only can continue as long as reads continuously made by the caller until EOF.
// When skipOnFail is set to true it'll skip the failed files and continue to process remaining ones.
func Download(files []File, w io.Writer, skipOnFail bool) error {
	zw := zip.NewWriter(w)
	defer zw.Close()

	download := func(f File) error {
		resp, err := DownloaderClient.Get(f.URL)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return errors.New("invalid file")
		}
		fw, err := zw.Create(f.Name)
		if err != nil {
			return err
		}
		_, err = io.Copy(fw, resp.Body)
		return err
	}

	for _, f := range files {
		if err := download(f); err != nil && !skipOnFail {
			return err
		}
	}
	return nil
}
