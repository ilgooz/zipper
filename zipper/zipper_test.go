package zipper

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	// serve test files so we can download them over HTTP.
	fs := http.FileServer(http.Dir("test-data"))
	ts := httptest.NewServer(fs)
	defer ts.Close()

	files := []File{
		{ts.URL + "/1.png", "1.png"},
		{ts.URL + "/2.jpeg", "2.jpeg"},
	}

	buf := new(bytes.Buffer)
	require.NoError(t, Download(files, buf, true))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	require.Len(t, zr.File, 2)
	require.Equal(t, "1.png", zr.File[0].Name)
	require.Equal(t, calcHash(t, zr.File[0]), "ad040e90865b0ba0966fba7ee9de0cd2b70b5c7317fd194156094099fe52bb73")
	require.Equal(t, "2.jpeg", zr.File[1].Name)
	require.Equal(t, calcHash(t, zr.File[1]), "cafd059538727a6352f9d923897f3e61378c34999a7a00d70e64f968b4ca7486")
}

func TestDownloadSkipFail(t *testing.T) {
	// serve test files so we can download them over HTTP.
	fs := http.FileServer(http.Dir("test-data"))
	ts := httptest.NewServer(fs)
	defer ts.Close()

	files := []File{
		{ts.URL + "/1.png", "1.png"},
		{ts.URL + "/not-exists.jpeg", "not-exists.jpeg"},
	}

	buf := new(bytes.Buffer)
	require.NoError(t, Download(files, buf, true))

	zr, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	require.NoError(t, err)

	require.Len(t, zr.File, 1)
	require.Equal(t, "1.png", zr.File[0].Name)
	require.Equal(t, calcHash(t, zr.File[0]), "ad040e90865b0ba0966fba7ee9de0cd2b70b5c7317fd194156094099fe52bb73")
}

func TestDownloadNoSkipFail(t *testing.T) {
	// serve test files so we can download them over HTTP.
	fs := http.FileServer(http.Dir("test-data"))
	ts := httptest.NewServer(fs)
	defer ts.Close()

	files := []File{
		{ts.URL + "/1.png", "1.png"},
		{ts.URL + "/not-exists.jpeg", "not-exists.jpeg"},
	}

	buf := new(bytes.Buffer)
	require.Equal(t, "invalid file", Download(files, buf, false).Error())
}

// calcHash calculates sha256 for zip file f.
func calcHash(t *testing.T, f *zip.File) string {
	r, err := f.Open()
	require.NoError(t, err)
	h := sha256.New()
	_, err = io.Copy(h, r)
	require.NoError(t, err)
	return fmt.Sprintf("%x", h.Sum(nil))
}
