package httpclient_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptrace"
	"sync"
	"sync/atomic"
)

// servedBody is the response body served by the non-pooling test servers.
const servedBody = "manifest-bytes"

// countingListener wraps a loopback listener and counts accepted
// connections, so tests can observe how many TCP connections were opened.
type countingListener struct {
	net.Listener

	accepted atomic.Int64
}

// newCountingListener opens a loopback listener for connection counting.
func newCountingListener(ctx context.Context) (*countingListener, error) {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("open loopback listener: %w", err)
	}

	return &countingListener{Listener: listener}, nil
}

// Accept counts every accepted connection.
func (listener *countingListener) Accept() (net.Conn, error) {
	connection, err := listener.Listener.Accept()
	if err != nil {
		return nil, fmt.Errorf("accept connection: %w", err)
	}

	listener.accepted.Add(1)

	return connection, nil
}

// requestObservation records what a test server saw for one request.
type requestObservation struct {
	closeRequested bool
}

// recordingHandler serves a fixed body while recording request observations
// under a mutex, so test goroutines can read them safely after requests
// complete.
type recordingHandler struct {
	mu           sync.Mutex
	observations []requestObservation
}

// ServeHTTP records the request and answers with the fixed test body.
func (handler *recordingHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	handler.mu.Lock()
	handler.observations = append(handler.observations, requestObservation{
		closeRequested: request.Close,
	})
	handler.mu.Unlock()

	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(servedBody))
}

// snapshot returns the recorded request observations.
func (handler *recordingHandler) snapshot() []requestObservation {
	handler.mu.Lock()
	defer handler.mu.Unlock()

	return append([]requestObservation(nil), handler.observations...)
}

// gotConnRecorder collects httptrace.GotConnInfo observations from client
// trace callbacks, which may run on transport goroutines.
type gotConnRecorder struct {
	mu    sync.Mutex
	conns []httptrace.GotConnInfo
}

// record stores one GotConnInfo observation.
func (recorder *gotConnRecorder) record(info httptrace.GotConnInfo) {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	recorder.conns = append(recorder.conns, info)
}

// snapshot returns the recorded GotConnInfo observations.
func (recorder *gotConnRecorder) snapshot() []httptrace.GotConnInfo {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	return append([]httptrace.GotConnInfo(nil), recorder.conns...)
}
