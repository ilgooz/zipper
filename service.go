package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/ilgooz/structionsite/x/xhttp"
	"github.com/ilgooz/structionsite/zipper"
)

// Service is a zipper service that implements algorithm specified in specs/ dir at root.
type Service struct {
	// Addr is the full network addr of the service.
	Addr string

	// Port is the port number that the service is listening at.
	Port int

	ln     net.Listener
	server *http.Server

	shutdownTimeout time.Duration
}

// New creates a new Service for given addr and graceful shutdownTimeout.
// It starts a tcp server that is ready to accept connections.
func New(addr string, shutdownTimeout time.Duration) (*Service, error) {
	s := &Service{
		shutdownTimeout: shutdownTimeout,
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	r.HandleFunc("/zip", handler).Methods("GET", "POST")

	s.ln = ln
	s.server = &http.Server{Handler: r}

	s.Port = ln.Addr().(*net.TCPAddr).Port
	s.Addr = ln.Addr().String()

	return s, nil
}

// GracefulStart starts the service with graceful shutdown option.
// When a quit signal is received, the service will stop receving new requests and will
// wait for the active ones to be completed until the timeout hits.
func (s *Service) GracefulStart() error {
	return xhttp.ServeGracefulServer("zip service", s.ln, s.server, s.shutdownTimeout)
}

// Close immediately terminates the Service.
func (s *Service) Close() error {
	defer s.ln.Close()
	return s.server.Shutdown(context.Background())
}

// handler handles incoming zip requests over HTTP.
func handler(w http.ResponseWriter, req *http.Request) {
	var files []zipper.File

	switch req.Method {
	case "GET":
		filesJSON := req.URL.Query().Get("files")
		if err := json.Unmarshal([]byte(filesJSON), &files); err != nil {
			xhttp.ResponseError(w, http.StatusBadRequest, err)
			return
		}
	case "POST":
		if err := xhttp.ParseJSON(req, &files); err != nil {
			xhttp.ResponseError(w, http.StatusBadRequest, err)
			return
		}
	}

	if err := validateFiles(files); err != nil {
		xhttp.ResponseError(w, http.StatusBadRequest, err)
		return
	}

	zipName := fmt.Sprintf("Files %s.zip", time.Now().Format("2006-01-02 at 15.04.05"))
	setZipHeaders(w, zipName)
	if err := zipper.Download(files, w, true); err != nil {
		panic(err)
	}
}

// setZipHeaders writes zip headers to HTTP response by using fileName as zip file's name.
func setZipHeaders(w http.ResponseWriter, fileName string) {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
}

// validateFiles validates files againts invalid urls and empty names.
func validateFiles(files []zipper.File) error {
	for i, f := range files {
		u, err := url.Parse(f.URL)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("invalid data at index %d: '%s' is not a valid url", i, f.URL)
		}
		if f.Name == "" {
			return fmt.Errorf("invalid data at index %d: file name cannot be empty", i)
		}
	}
	return nil
}
