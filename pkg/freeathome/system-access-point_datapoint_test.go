package freeathome

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/lion-and-bear/freeathome/v2/pkg/models"
)

func TestSystemAccessPointGetDatapoint(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       loadTestResponseBody(t, "get_datapoint.json"),
		Header:     make(http.Header),
	}

	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetDatapoint("abcd1234", "ch0000", "odp0001")

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
	expectedUrl := "https://localhost/fhapi/v1/api/rest/datapoint/00000000-0000-0000-0000-000000000000/abcd1234.ch0000.odp0001"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the result is not nil and contains the expected data
	if result == nil || *result == nil {
		t.Error("Expected non-nil result")
	}
	if len(*result) != 1 {
		t.Errorf("Expected the response to contain one datapoint, got %d", len(*result))
	}
	if len((*result)[models.EmptyUUID].Values) != 1 {
		t.Errorf("Expected 1 datapoint, got %d", len((*result)[models.EmptyUUID].Values))
	}
	if (*result)[models.EmptyUUID].Values[0] != "1" {
		t.Errorf("Expected datapoint value to be '1', got '%s'", (*result)[models.EmptyUUID].Values[0])
	}
}

func TestSystemAccessPointGetDatapointCallError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	error := errors.New("Test Error")
	roundtripper := &MockRoundTripper{
		Response: nil,
		Err:      error,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetDatapoint("abcd1234", "ch0000", "odp0001")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to get datapoint\"") ||
		!strings.Contains(logOutput, "error=\"Get \\\"https://localhost/fhapi/v1/api/rest/datapoint/00000000-0000-0000-0000-000000000000/abcd1234.ch0000.odp0001\\\": Test Error") {
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
	expected := "Get \"https://localhost/fhapi/v1/api/rest/datapoint/00000000-0000-0000-0000-000000000000/abcd1234.ch0000.odp0001\": Test Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointGetDatapointErrorResponse(t *testing.T) {
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

	result, err := sysAp.GetDatapoint("abcd1234", "ch0000", "odp0001")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to get datapoint\"") ||
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
	expectedUrl := "https://localhost/fhapi/v1/api/rest/datapoint/00000000-0000-0000-0000-000000000000/abcd1234.ch0000.odp0001"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the error message is correct
	expected := "failed to get datapoint: Internal Server Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointGetDatapointUnmarshalError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"00000000-0000-0000-0000-000000000000":{"values":[123]}}`)),
		Header:     make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.GetDatapoint("abcd1234", "ch0000", "odp0001")

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
	expected := "json: cannot unmarshal number into Go struct field GetDataPoint.values of type string"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointSetDatapoint(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       loadTestResponseBody(t, "set_datapoint.json"),
		Header:     make(http.Header),
	}

	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.SetDatapoint("abcd1234", "ch0000", "idp0001", "123")

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
	expectedUrl := "https://localhost/fhapi/v1/api/rest/datapoint/00000000-0000-0000-0000-000000000000/abcd1234.ch0000.idp0001"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the result is not nil and contains the expected data
	if result == nil || *result == nil {
		t.Error("Expected non-nil result")
	}
	if len(*result) != 1 {
		t.Errorf("Expected the response to contain one set datapoint response, got %d", len(*result))
	}
	val, ok := (*result)[models.EmptyUUID]["result"]
	if !ok {
		t.Error("Expected the set datapoint to contain a result")
	}
	if val != "OK" {
		t.Error("Expected the set datapoint result to be 'OK'")
	}
}

func TestSystemAccessPointSetDatapointCallError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	error := errors.New("Test Error")
	roundtripper := &MockRoundTripper{
		Response: nil,
		Err:      error,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.SetDatapoint("abcd1234", "ch0000", "idp0001", "123")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to set datapoint\"") ||
		!strings.Contains(logOutput, "error=\"Put \\\"https://localhost/fhapi/v1/api/rest/datapoint/00000000-0000-0000-0000-000000000000/abcd1234.ch0000.idp0001\\\": Test Error") {
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
	expected := "Put \"https://localhost/fhapi/v1/api/rest/datapoint/00000000-0000-0000-0000-000000000000/abcd1234.ch0000.idp0001\": Test Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointSetDatapointErrorResponse(t *testing.T) {
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

	result, err := sysAp.SetDatapoint("abcd1234", "ch0000", "idp0001", "123")

	// Check if the log output contains the expected error message
	logOutput := buf.String()
	if !strings.Contains(logOutput, "msg=\"failed to set datapoint\"") ||
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
	expectedUrl := "https://localhost/fhapi/v1/api/rest/datapoint/00000000-0000-0000-0000-000000000000/abcd1234.ch0000.idp0001"
	if roundtripper.Request.URL.String() != expectedUrl {
		t.Errorf("Expected URL '%s', got '%s'", expectedUrl, roundtripper.Request.URL.String())
	}

	// Check if the error message is correct
	expected := "failed to set datapoint: Internal Server Error"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}

func TestSystemAccessPointSetDatapointUnmarshalError(t *testing.T) {
	sysAp, buf, _ := setupSysAp(t, true, false)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"00000000-0000-0000-0000-000000000000":{"result": 1}}`)),
		Header:     make(http.Header),
	}
	roundtripper := &MockRoundTripper{
		Response: response,
		Err:      nil,
	}
	sysAp.config.Client.SetTransport(roundtripper)

	result, err := sysAp.SetDatapoint("abcd1234", "ch0000", "idp0001", "123")

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
	expected := "json: cannot unmarshal number into Go value of type string"
	if err.Error() != expected {
		t.Errorf(expectedErrorGotValue, expected, err)
	}
}
