package freeathome

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/lion-and-bear/freeathome/v2/pkg/models"
)

func TestSystemAccessPointGetDevice(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       loadTestResponseBody(t, "device.json"),
		Header:     make(http.Header),
	}

	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetDevice("600028E1ED13")

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
	expectedUrl := "https://localhost/fhapi/v1/api/rest/device/00000000-0000-0000-0000-000000000000/600028E1ED13"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the result is not nil and contains the expected data
	if result == nil || *result == nil {
		t.Error("Expected non-nil result")
	}
	if len(*result) != 1 {
		t.Errorf("Expected the response to contain one device response, got %d", len(*result))
	}
	if len((*result)[models.EmptyUUID].Devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len((*result)[models.EmptyUUID].Devices))
	}
	if *(*result)[models.EmptyUUID].Devices["600028E1ED13"].NativeID != "47110815AA" {
		t.Errorf("Expected created virtual device serial to be '47110815AA', got '%s'", *(*result)[models.EmptyUUID].Devices["600028E1ED13"].NativeID)
	}
}

func TestSystemAccessPointGetDeviceCallError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	error := errors.New("Test Error")
	roundtripper := &MockRoundTripper{
		Response: nil,
		Err:      error,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetDevice("600028E1ED13")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to get device\"") ||
		!strings.Contains(logOutput, "error=\"Get \\\"https://localhost/fhapi/v1/api/rest/device/00000000-0000-0000-0000-000000000000/600028E1ED13\\\": Test Error") {
		t.Errorf(unexpectedLogOutput, logOutput)
	}

	// Check if result is nil and error is not nil
	if result != nil {
		t.Errorf("Expected result to be nil, got %v", result)
	}
	if err == nil {
		t.Fatal(expectedErrorGotNil)
	}

	// Check if the error message is correct
	expected := "Get \"https://localhost/fhapi/v1/api/rest/device/00000000-0000-0000-0000-000000000000/600028E1ED13\": Test Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointGetDeviceErrorResponse(t *testing.T) {
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

	result, err := sysAp.GetDevice("600028E1ED13")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to get device\"") ||
		!strings.Contains(logOutput, "level=ERROR") ||
		!strings.Contains(logOutput, "status=\"Internal Server Error\"") ||
		!strings.Contains(logOutput, "body=\"Internal Server Error\"") {
		t.Errorf(unexpectedLogOutput, logOutput)
	}

	// Check if result is nil and error is not nil
	if result != nil {
		t.Error(expectedNil)
	}
	if err == nil {
		t.Error(expectedErrorGotNil)
	}

	// Check if the request method and URL are correct
	if roundtripper.Request.Method != http.MethodGet {
		t.Errorf("Expected GET request, got %s", roundtripper.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/device/00000000-0000-0000-0000-000000000000/600028E1ED13"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the error message is correct
	expected := "failed to get device: Internal Server Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointGetDeviceUnmarshalError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"00000000-0000-0000-0000-000000000000":{"devices":{"abcd12345":{"nativeId": 123}}}}`)),
		Header:     make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetDevice("600028E1ED13")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to parse response body\"") ||
		!strings.Contains(logOutput, "level=ERROR") {
		t.Errorf(unexpectedLogOutput, logOutput)
	}

	// Check if result is nil and error is not nil
	if result != nil {
		t.Error(expectedNil)
	}
	if err == nil {
		t.Fatal(expectedErrorGotNil)
	}

	// Check if the error message is correct
	expected := "json: cannot unmarshal number into Go struct field Device.devices.nativeId of type string"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}
