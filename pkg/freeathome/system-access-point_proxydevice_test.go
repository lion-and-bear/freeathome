package freeathome

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/lion-and-bear/freeathome/v2/pkg/models"
)

func TestSystemAccessPointTriggerProxyDevice(t *testing.T) {
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

	result, err := sysAp.TriggerProxyDevice("doorring", "600028E1ED13", "shortpress")

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
	expectedUrl := "https://localhost/fhapi/v1/api/rest/proxydevice/00000000-0000-0000-0000-000000000000/doorring/600028E1ED13/action/shortpress"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the result is not nil and contains the expected data
	if result == nil || *result == nil {
		t.Error("Expected non-nil result")
	}
	if len(*result) != 1 {
		t.Errorf("Expected the response to contain one proxy device response, got %d", len(*result))
	}
	if len((*result)[models.EmptyUUID].Devices) != 1 {
		t.Errorf("Expected 1 proxy device, got %d", len((*result)[models.EmptyUUID].Devices))
	}
	if *(*result)[models.EmptyUUID].Devices["600028E1ED13"].NativeID != "47110815AA" {
		t.Errorf("Expected proxy device native ID to be '47110815AA', got '%s'", *(*result)[models.EmptyUUID].Devices["600028E1ED13"].NativeID)
	}
}

func TestSystemAccessPointTriggerProxyDeviceCallError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	error := errors.New("Test Error")
	roundtripper := &MockRoundTripper{
		Response: nil,
		Err:      error,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.TriggerProxyDevice("doorring", "600028E1ED13", "shortpress")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to trigger proxy device\"") ||
		!strings.Contains(logOutput, "error=\"Get \\\"https://localhost/fhapi/v1/api/rest/proxydevice/00000000-0000-0000-0000-000000000000/doorring/600028E1ED13/action/shortpress\\\": Test Error") {
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
	expected := "Get \"https://localhost/fhapi/v1/api/rest/proxydevice/00000000-0000-0000-0000-000000000000/doorring/600028E1ED13/action/shortpress\": Test Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointTriggerProxyDeviceErrorResponse(t *testing.T) {
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

	result, err := sysAp.TriggerProxyDevice("doorring", "600028E1ED13", "shortpress")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to trigger proxy device\"") ||
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
	expectedUrl := "https://localhost/fhapi/v1/api/rest/proxydevice/00000000-0000-0000-0000-000000000000/doorring/600028E1ED13/action/shortpress"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the error message is correct
	expected := "failed to trigger proxy device: Internal Server Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointTriggerProxyDeviceUnmarshalError(t *testing.T) {
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

	result, err := sysAp.TriggerProxyDevice("doorring", "600028E1ED13", "shortpress")

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

func TestSystemAccessPointSetProxyDeviceValue(t *testing.T) {
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

	result, err := sysAp.SetProxyDeviceValue("doorring", "600028E1ED13", "123")

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
	if roundtripper.Request.Method != http.MethodPut {
		t.Errorf("Expected PUT request, got %s", roundtripper.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/proxydevice/00000000-0000-0000-0000-000000000000/doorring/600028E1ED13/value/123"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}
	// Check if the result is not nil and contains the expected data
	if result == nil || *result == nil {
		t.Error("Expected non-nil result")
	}
	if len(*result) != 1 {
		t.Errorf("Expected the response to contain one proxy device response, got %d", len(*result))
	}
	if len((*result)[models.EmptyUUID].Devices) != 1 {
		t.Errorf("Expected 1 proxy device, got %d", len((*result)[models.EmptyUUID].Devices))
	}
	if *(*result)[models.EmptyUUID].Devices["600028E1ED13"].NativeID != "47110815AA" {
		t.Errorf("Expected proxy device native ID to be '47110815AA', got '%s'", *(*result)[models.EmptyUUID].Devices["600028E1ED13"].NativeID)
	}
}

func TestSystemAccessPointSetProxyDeviceValueCallError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	error := errors.New("Test Error")
	roundtripper := &MockRoundTripper{
		Response: nil,
		Err:      error,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.SetProxyDeviceValue("doorring", "600028E1ED13", "123")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to set proxy device value\"") ||
		!strings.Contains(logOutput, "error=\"Put \\\"https://localhost/fhapi/v1/api/rest/proxydevice/00000000-0000-0000-0000-000000000000/doorring/600028E1ED13/value/123\\\": Test Error") {
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
	expected := "Put \"https://localhost/fhapi/v1/api/rest/proxydevice/00000000-0000-0000-0000-000000000000/doorring/600028E1ED13/value/123\": Test Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointSetProxyDeviceValueErrorResponse(t *testing.T) {
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

	result, err := sysAp.SetProxyDeviceValue("doorring", "600028E1ED13", "123")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to set proxy device value\"") ||
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
	if roundtripper.Request.Method != http.MethodPut {
		t.Errorf("Expected PUT request, got %s", roundtripper.Request.Method)
	}
	expectedUrl := "https://localhost/fhapi/v1/api/rest/proxydevice/00000000-0000-0000-0000-000000000000/doorring/600028E1ED13/value/123"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the error message is correct
	expected := "failed to set proxy device value: Internal Server Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointSetProxyDeviceValueUnmarshalError(t *testing.T) {
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

	result, err := sysAp.SetProxyDeviceValue("doorring", "600028E1ED13", "123")

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
