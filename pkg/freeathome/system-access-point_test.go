package freeathome

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// TestSystemAccessPointDefaultLogger tests the default logger functionality of SystemAccessPoint.
func TestSystemAccessPointDefaultLogger(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer
	handler := slog.NewTextHandler(&buf, nil)
	slog.SetDefault(slog.New(handler))

	// Create a SystemAccessPoint with the default logger
	NewSystemAccessPointWithDefaults("localhost", "user", "password")

	// Check if the log output contains the expected message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "No logger provided for SystemAccessPoint. Using default logger.") {
		t.Errorf("Expected log output to contain 'No logger provided for SystemAccessPoint. Using default logger.', got: %s", logOutput)
	}
}

// TestNoConfigErrors tests that an error is returned when a nil config is passed to NewSystemAccessPoint.
func TestNoConfigErrors(t *testing.T) {
	sysap, err := NewSystemAccessPoint(nil)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if sysap != nil {
		t.Errorf("Expected nil, got %v", sysap)
	}
}

// TestMustNewSystemAccessPoint tests that MustNewSystemAccessPoint panics when a nil config is passed to NewSystemAccessPoint.
func TestMustNewSystemAccessPoint(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("Expected panic, got nil")
		}
	}()
	MustNewSystemAccessPoint(nil)
}

func TestSystemAccessPointWebSocketRawOutputGetter(t *testing.T) {
	sysAp, _, _ := setupSysAp(t, true, false)
	if sysAp.GetWebSocketRawOutput() != nil {
		t.Fatal("expected nil before set")
	}
	var buf bytes.Buffer
	sysAp.SetWebSocketRawOutput(&buf)
	if sysAp.GetWebSocketRawOutput() != &buf {
		t.Fatal("getter should return configured writer")
	}
	sysAp.SetWebSocketRawOutput(nil)
	if sysAp.GetWebSocketRawOutput() != nil {
		t.Fatal("expected nil after clear")
	}
}

// TestSystemAccessPointGetHostName tests the GetHostName method of SystemAccessPoint.
func TestSystemAccessPointGetHostName(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	expected := "localhost"

	actual := sysAp.GetHostName()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual host name matches the expected host name
	if actual != expected {
		t.Errorf("Expected host name '%s', got '%s'", expected, actual)
	}
}

// TestSystemAccessPointGetTlsEnabled tests the GetTlsEnabled method of SystemAccessPoint.
func TestSystemAccessPointGetTlsEnabled(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	expected := true

	actual := sysAp.GetTlsEnabled()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual TLS enabled status matches the expected status
	if actual != expected {
		t.Errorf("Expected TLS enabled '%v', got '%v'", expected, actual)
	}
}

// TestSystemAccessPointGetSkipTLSVerify tests the GetSkipTLSVerify method of SystemAccessPoint.
func TestSystemAccessPointGetSkipTLSVerify(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	expected := false

	actual := sysAp.GetSkipTLSVerify()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual skip TLS verify status matches the expected status
	if actual != expected {
		t.Errorf("Expected skip TLS verify '%v', got '%v'", expected, actual)
	}
}

// TestSystemAccessPointGetSkipTLSVerifyEnabled tests the GetSkipTLSVerify method of SystemAccessPoint when skip TLS verify is enabled.
func TestSystemAccessPointGetSkipTLSVerifyEnabled(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, true)
	expected := true

	actual := sysAp.GetSkipTLSVerify()

	// Check if the log output is empty
	logOutput := buf.String()
	if !strings.Contains(logOutput, "this is not recommended") {
		t.Errorf("Expected log output to contain 'this is not recommended', got: %s", logOutput)
	}

	// Check if the actual skip TLS verify status matches the expected status
	if actual != expected {
		t.Errorf("Expected skip TLS verify '%v', got '%v'", expected, actual)
	}
}

// TestSystemAccessPointGetVerboseErrors tests the GetVerboseErrors method of SystemAccessPoint.
func TestSystemAccessPointGetVerboseErrors(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	expected := false

	actual := sysAp.GetVerboseErrors()

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual verbose errors status matches the expected status
	if actual != expected {
		t.Errorf("Expected verbose errors '%v', got '%v'", expected, actual)
	}
}

// TestSystemAccessPointGetUrlWithoutTls tests the GetUrl method of SystemAccessPoint without TLS.
func TestSystemAccessPointGetUrlWithoutTls(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, false, false)
	expected := "http://localhost/fhapi/v1/api/rest/test123"
	actual := sysAp.GetUrl("test123")

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, actual)
	}
}

// TestSystemAccessPointGetUrlWithTls tests the GetUrl method of SystemAccessPoint with TLS.
func TestSystemAccessPointGetUrlWithTls(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)

	actual := sysAp.GetUrl("test123")
	expected := "https://localhost/fhapi/v1/api/rest/test123"

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the actual URL matches the expected URL
	if actual != expected {
		t.Errorf("Expected URL '%s', got '%s'", expected, actual)
	}
}
