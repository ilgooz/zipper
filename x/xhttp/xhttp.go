package xhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// ParseJSON parses json payload from request's payload into o.
func ParseJSON(r *http.Request, o interface{}) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(o)
}

// ServeGracefulServer starts http server s with given shutdown timeout and name by using listener ln.
// Once quit signal is received for the process, server stops accepting new requests and waits to finish on
// going responses until the timeout occurs.
func ServeGracefulServer(name string, ln net.Listener, s *http.Server, timeout time.Duration) error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	serveErr := make(chan error)
	go func() {
		if err := s.Serve(ln); err != nil && err != http.ErrServerClosed {
			serveErr <- fmt.Errorf("%s listen: %s\n", name, err)
		}
	}()
	log.Printf("%s has started at %s", name, ln.Addr().String())

	select {
	case <-done:
		log.Printf("%s is shutting down...", name)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			log.Fatalf("%s shutdown failed: %+v", name, err)
		}
		log.Printf("%s shutdown", name)
		return nil
	case err := <-serveErr:
		return err
	}
}

type ErrorResponseBody struct {
	Error ErrorResponse `json:"error"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

// ResponseError responses a json error message with given http status.
// Example output: 400 { "error": { "message": "an error occurred" } }
func ResponseError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	data, _ := json.Marshal(ErrorResponseBody{
		Error: ErrorResponse{Message: err.Error()},
	})
	w.Write(data)
}
