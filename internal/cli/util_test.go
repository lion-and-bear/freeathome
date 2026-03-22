package cli

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"log/slog"

	"github.com/go-resty/resty/v2"
	"github.com/lion-and-bear/freeathome/v2/pkg/freeathome"
	"github.com/spf13/viper"
)

// setupViper creates a viper instance with test configuration
func setupViper(t *testing.T) *viper.Viper {
	t.Helper()

	// Create temporary directory for test
	configFileDir = t.TempDir()

	// Create config file in the expected location (~/.freeathome/config.yaml)
	configDir := filepath.Join(configFileDir, ".freeathome")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "config.yaml")

	// Create a minimal config file
	configData := `hostname: test-host
username: test-user
password: test-pass`

	err := os.WriteFile(configFile, []byte(configData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	return v
}

// setupMock creates a mock SystemAccessPoint with the given response
func setupMock(t *testing.T, v *viper.Viper, responseCode int, responseBody string) (*freeathome.SystemAccessPoint, *bytes.Buffer, chan slog.Record) {
	t.Helper()

	// Create mock response
	response := &http.Response{
		StatusCode: responseCode,
		Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
		Header:     make(http.Header),
	}

	// Create mock transport
	mockTransport := &MockRoundTripper{
		Response: response,
	}

	// Create resty client with mock transport
	client := resty.New().SetTransport(mockTransport)

	// Create logger
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := freeathome.NewDefaultLogger(handler)

	// Create SystemAccessPoint with mock client
	config := freeathome.NewConfig(v.GetString("hostname"), v.GetString("username"), v.GetString("password"))
	config.TLSEnabled = true
	config.Logger = logger
	config.Client = client

	return freeathome.MustNewSystemAccessPoint(config), nil, nil
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

// createTestConfigFile creates a test config file
func createTestConfigFile(t *testing.T, configData string) string {
	t.Helper()

	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	if configData != "" {
		err := os.WriteFile(configFile, []byte(configData), 0644)
		if err != nil {
			t.Fatalf("Failed to create test config file: %v", err)
		}
	}

	return configFile
}

// createViperWithConfig creates a viper instance with config file
func createViperWithConfig(t *testing.T, configData string) *viper.Viper {
	t.Helper()

	configFile := createTestConfigFile(t, configData)
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")

	return v
}

// assertConfigValues asserts config values
func assertConfigValues(t *testing.T, cfg *Config, expectedHostname, expectedUsername, expectedPassword string) {
	t.Helper()

	if cfg.Hostname != expectedHostname {
		t.Errorf("Expected Hostname to be '%s', got '%s'", expectedHostname, cfg.Hostname)
	}

	if cfg.Username != expectedUsername {
		t.Errorf("Expected Username to be '%s', got '%s'", expectedUsername, cfg.Username)
	}

	if cfg.Password != expectedPassword {
		t.Errorf("Expected Password to be '%s', got '%s'", expectedPassword, cfg.Password)
	}
}
