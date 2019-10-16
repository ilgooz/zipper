package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ilgooz/structionsite/x/xhttp"
	"github.com/stretchr/testify/require"
)

func TestZipWithPost(t *testing.T) {
	// serve test files via HTTP.
	fs := http.FileServer(http.Dir("zipper/test-data"))
	ts := httptest.NewServer(fs)
	defer ts.Close()

	// start the zipper service.
	s, err := New(":0", time.Minute)
	require.NoError(t, err)
	defer s.Close()
	go s.GracefulStart()

	resp, err := http.Post(genServiceURL(s.Port), "application/json", bytes.NewBuffer([]byte(
		fmt.Sprintf(`[
			{"url":"%s/1.png", "filename":"1.png"},
			{"url":"%s/2.jpeg", "filename":"2.jpeg"}
		]`, ts.URL, ts.URL),
	)))
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	require.Len(t, zr.File, 2)
	require.Equal(t, "1.png", zr.File[0].Name)
	require.Equal(t, calcHash(t, zr.File[0]), "ad040e90865b0ba0966fba7ee9de0cd2b70b5c7317fd194156094099fe52bb73")
	require.Equal(t, "2.jpeg", zr.File[1].Name)
	require.Equal(t, calcHash(t, zr.File[1]), "cafd059538727a6352f9d923897f3e61378c34999a7a00d70e64f968b4ca7486")
}

func TestZipWithOneBrokenFile(t *testing.T) {
	// serve test files via HTTP.
	fs := http.FileServer(http.Dir("zipper/test-data"))
	ts := httptest.NewServer(fs)
	defer ts.Close()

	// start the zipper service.
	s, err := New(":0", time.Minute)
	require.NoError(t, err)
	defer s.Close()
	go s.GracefulStart()

	resp, err := http.Post(genServiceURL(s.Port), "application/json", bytes.NewBuffer([]byte(
		fmt.Sprintf(`[
			{"url":"%s/not-exists.png", "filename":"not-exists.png"},
			{"url":"%s/2.jpeg", "filename":"2.jpeg"}
		]`, ts.URL, ts.URL),
	)))
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	require.Len(t, zr.File, 1)
	require.Equal(t, "2.jpeg", zr.File[0].Name)
	require.Equal(t, calcHash(t, zr.File[0]), "cafd059538727a6352f9d923897f3e61378c34999a7a00d70e64f968b4ca7486")
}

func TestZipWithGet(t *testing.T) {
	// serve test files via HTTP.
	fs := http.FileServer(http.Dir("zipper/test-data"))
	ts := httptest.NewServer(fs)
	defer ts.Close()

	// start the zipper service.
	s, err := New(":0", time.Minute)
	require.NoError(t, err)
	defer s.Close()
	go s.GracefulStart()

	u, _ := url.Parse(genServiceURL(s.Port))
	q := u.Query()
	q.Set("files", fmt.Sprintf(`[
		{"url":"%s/1.png", "filename":"1.png"},
		{"url":"%s/2.jpeg", "filename":"2.jpeg"}
	]`, ts.URL, ts.URL))
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)

	require.Len(t, zr.File, 2)
	require.Equal(t, "1.png", zr.File[0].Name)
	require.Equal(t, calcHash(t, zr.File[0]), "ad040e90865b0ba0966fba7ee9de0cd2b70b5c7317fd194156094099fe52bb73")
	require.Equal(t, "2.jpeg", zr.File[1].Name)
	require.Equal(t, calcHash(t, zr.File[1]), "cafd059538727a6352f9d923897f3e61378c34999a7a00d70e64f968b4ca7486")
}

func TestZipInvalidURL(t *testing.T) {
	// start the zipper service.
	s, err := New(":0", time.Minute)
	require.NoError(t, err)
	defer s.Close()
	go s.GracefulStart()

	resp, err := http.Post(genServiceURL(s.Port), "application/json", bytes.NewBuffer([]byte(`[
		{"url":"invalid-url", "filename":"1.png"}
	]`)))
	require.NoError(t, err)
	defer resp.Body.Close()

	var errResp xhttp.ErrorResponseBody
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "invalid data at index 0: 'invalid-url' is not a valid url", errResp.Error.Message)
}

func TestZipInvalidFileName(t *testing.T) {
	// start the zipper service.
	s, err := New(":0", time.Minute)
	require.NoError(t, err)
	defer s.Close()
	go s.GracefulStart()

	resp, err := http.Post(genServiceURL(s.Port), "application/json", bytes.NewBuffer([]byte(`[
		{"url":"http://localhost/test", "filename":""}
	]`)))
	require.NoError(t, err)
	defer resp.Body.Close()

	var errResp xhttp.ErrorResponseBody
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&errResp))
	require.Equal(t, "invalid data at index 0: file name cannot be empty", errResp.Error.Message)
}

func genServiceURL(port int) string {
	return fmt.Sprintf("http://localhost:%d/zip", port)
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
