package freeathome

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/lion-and-bear/freeathome/v2/pkg/models"
)

// TestSystemAccessPointGetConfiguration tests the GetConfiguration method of SystemAccessPoint.
func TestSystemAccessPointGetConfiguration(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       loadTestResponseBody(t, "configuration.json"),
		Header:     make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetConfiguration()

	// Check for errors
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check if the log output is empty
	logOutput := buf.String()
	if logOutput != "" {
		t.Errorf("Expected no log output, got: %s", logOutput)
	}

	// Check if the request method and URL are correct
	if roundtripper.Request.Method != http.MethodGet {
		t.Errorf("Expected GET request, got %s", roundtripper.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/configuration"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the result is not nil and contains the expected data
	if *result == nil {
		t.Error("Expected non-nil result")
	}
	if len(*result) != 1 {
		t.Errorf("Expected the configuration to contain one system access point, got %d", len(*result))
	}
	if len((*result)[models.EmptyUUID].Devices) != 76 {
		t.Errorf("Expected 76 devices, got %d", len((*result)[models.EmptyUUID].Devices))
	}
}

// TestSystemAccessPointGetConfigurationCallError tests the GetConfiguration method of SystemAccessPoint
func TestSystemAccessPointGetConfigurationCallError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	error := errors.New("Test Error")
	roundtripper := &MockRoundTripper{
		Response: nil,
		Err:      error,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetConfiguration()

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to get configuration\"") ||
		!strings.Contains(logOutput, "error=\"Get \\\"https://localhost/fhapi/v1/api/rest/configuration\\\": Test Error\"") {
		t.Errorf(unexpectedLogOutput, logOutput)
	}
	// Check if result is nil and error is not nil
	if result != nil {
		t.Error(expectedNil)
	}
	if err == nil {
		t.Error(expectedErrorGotNil)
	}

	// Check if the error message is correct
	expected := "Get \"https://localhost/fhapi/v1/api/rest/configuration\": Test Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

// TestSystemAccessPointGetConfigurationErrorResponse tests the GetConfiguration method of SystemAccessPoint
func TestSystemAccessPointGetConfigurationErrorResponse(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	response := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Status:     "Internal Server Error",
		Body:       io.NopCloser(strings.NewReader("Internal Server Error")),
		Header:     make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetConfiguration()

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to get configuration\"") ||
		!strings.Contains(logOutput, "level=ERROR") ||
		!strings.Contains(logOutput, "status=\"Internal Server Error\"") ||
		!strings.Contains(logOutput, "body=\"Internal Server Error\"") {
		t.Errorf(unexpectedLogOutput, logOutput)
	}

	// Check if result is nil and error is not nil
	if err == nil {
		t.Fatal(expectedErrorGotNil)
	}
	if result != nil {
		t.Error(expectedNil)
	}

	// Check if the request method and URL are correct
	if roundtripper.Request.Method != http.MethodGet {
		t.Errorf("Expected GET request, got %s", response.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/configuration"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the error message is correct
	expected := "failed to get configuration: Internal Server Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

// TestSystemAccessPointGetConfigurationUnmarshalError tests the GetConfiguration method of SystemAccessPoint
func TestSystemAccessPointGetConfigurationUnmarshalError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	response := &http.Response{
		StatusCode: http.StatusOK,
		// This is intentionally malformed JSON to trigger the unmarshal error
		Body:   io.NopCloser(strings.NewReader("{\"devices\": [{\"id\": \"device1\"}, {\"id\": \"device2\"}]}")),
		Header: make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetConfiguration()

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to parse response body\"") ||
		!strings.Contains(logOutput, "level=ERROR") {
		t.Errorf(unexpectedLogOutput, logOutput)
	}

	// Check if result is nil and error is not nil
	if err == nil {
		t.Fatal(expectedErrorGotNil)
	}
	if result != nil {
		t.Error(expectedNil)
	}

	// Check if the error message is correct
	expected := "json: cannot unmarshal array into Go value of type models.SysAP"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}
