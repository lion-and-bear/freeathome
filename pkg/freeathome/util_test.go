package freeathome

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// ThreadSafeBuffer wraps bytes.Buffer with thread-safe operations
type ThreadSafeBuffer struct {
	buf bytes.Buffer
	mu  sync.RWMutex
}

func (tb *ThreadSafeBuffer) Write(p []byte) (n int, err error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.buf.Write(p)
}

func (tb *ThreadSafeBuffer) String() string {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	return tb.buf.String()
}

func (tb *ThreadSafeBuffer) Reset() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.buf.Reset()
}

// fakeClock is a mock implementation of the clock interface for testing purposes.
type fakeClock struct {
	now        time.Time
	afterCalls []time.Duration
	mu         sync.Mutex
}

func (mt *fakeClock) Now() time.Time {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	return mt.now
}

func (mt *fakeClock) After(d time.Duration) <-chan time.Time {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.afterCalls = append(mt.afterCalls, d)
	ch := make(chan time.Time, 1)
	ch <- mt.now.Add(d)
	close(ch)
	return ch
}

func (mt *fakeClock) Sleep(d time.Duration) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.now = mt.now.Add(d)
}

const expectedErrorGotNil = "Expected error, got nil"
const expectedErrorGotValue = "Expected error '%s', got '%v'"
const expectedNil = "Expected nil result"
const unexpectedLogOutput = "Unexpected log output, got: %s"

// assertUnmarshalTypeError checks that err is a *json.UnmarshalTypeError for a JSON value of kind value
// that could not be assigned to a Go value of type T. The error message itself is not compared, since its
// wording differs between Go versions.
func assertUnmarshalTypeError[T any](t *testing.T, err error, value string) {
	t.Helper()

	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		t.Fatalf("Expected *json.UnmarshalTypeError, got %T: %v", err, err)
	}
	if typeErr.Value != value {
		t.Errorf("Expected JSON value '%s', got '%s'", value, typeErr.Value)
	}
	if expected := reflect.TypeFor[T](); typeErr.Type != expected {
		t.Errorf("Expected Go type '%v', got '%v'", expected, typeErr.Type)
	}
}

// setupSysAp initializes a SystemAccessPoint with a mock logger and returns it along with a buffer to capture log output.
func setupSysAp(t *testing.T, tlsEnabled bool, skipTLSVerify bool) (*SystemAccessPoint, *ThreadSafeBuffer, chan slog.Record) {
	t.Helper()

	// Create a thread-safe buffer to capture log output
	var buf ThreadSafeBuffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	// Create a channel handler to capture log records
	channelHandler := &ChannelHandler{
		next:    handler,
		records: make(chan slog.Record, 100),
	}
	// Create the logger
	logger := NewDefaultLogger(channelHandler)

	// Create a SystemAccessPoint with the default logger
	config := NewConfig("localhost", "user", "password")
	config.TLSEnabled = tlsEnabled
	config.SkipTLSVerify = skipTLSVerify
	config.Logger = logger
	return MustNewSystemAccessPoint(config), &buf, channelHandler.records
}

// setupSysApWebSocket initializes a SystemAccessPointWebSocket with a mock logger and returns it along with a buffer to capture log output.
func setupSysApWebSocket(t *testing.T, tlsEnabled bool, skipTLSVerify bool) (*SystemAccessPointWebSocket, *ThreadSafeBuffer, chan slog.Record) {
	t.Helper()

	// Create a thread-safe buffer to capture log output
	var buf ThreadSafeBuffer
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	// Create a channel handler to capture log records
	channelHandler := &ChannelHandler{
		next:    handler,
		records: make(chan slog.Record, 100),
	}
	// Create the logger
	logger := NewDefaultLogger(channelHandler)

	// Create a SystemAccessPoint with the default logger
	config := NewConfig("localhost", "user", "password")
	config.TLSEnabled = tlsEnabled
	config.SkipTLSVerify = skipTLSVerify
	config.Logger = logger
	sysAp := MustNewSystemAccessPoint(config)

	// Create a SystemAccessPointWebSocket
	return &SystemAccessPointWebSocket{
		sysAp:     sysAp,
		waitGroup: sync.WaitGroup{},
	}, &buf, channelHandler.records
}

// MockRoundTripper is a mock implementation of http.RoundTripper for testing purposes.
type MockRoundTripper struct {
	Request  *http.Request
	Response *http.Response
	Err      error
}

// RoundTrip executes a single HTTP transaction and returns the response.
func (m *MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	m.Request = req
	return m.Response, m.Err
}

// It captures the request and response for testing purposes.
func loadTestResponseBody(t *testing.T, filename string) io.ReadCloser {
	t.Helper()

	path := filepath.Join("..", "..", "testdata", filename)
	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("failed to read test file %s: %v", filename, err)
	}

	content := string(data)
	return io.NopCloser(strings.NewReader(content))
}
