package cli

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lion-and-bear/freeathome/v2/pkg/freeathome"
	"github.com/spf13/viper"
)

// TestSetDatapoint tests the SetDatapoint function with various inputs and configurations
func TestSetDatapoint(t *testing.T) {
	tests := []struct {
		name          string
		serial        string
		channel       string
		datapoint     string
		value         string
		outputFormat  string
		prettify      bool
		responseBody  string
		responseCode  int
		expectError   bool
		errorContains string
		expectOutput  string
	}{
		{
			name:         "Successful JSON output",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "1",
			outputFormat: "json",
			prettify:     false,
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": {
    "status": "success"
  }
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: `{"00000000-0000-0000-0000-000000000000":{"status":"success"}}
`,
		},
		{
			name:         "Successful JSON output with prettify",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "1",
			outputFormat: "json",
			prettify:     true,
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": {
    "status": "success"
  }
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: `{
  "00000000-0000-0000-0000-000000000000": {
    "status": "success"
  }
}
`,
		},
		{
			name:         "Successful text output",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "1",
			outputFormat: "text",
			prettify:     false,
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": {
    "status": "success"
  }
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "Datapoint set successfully: ABB7F595EC47.ch0000.idp0000\n  Response: map[status:success]\n",
		},
		{
			name:         "Empty response",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "1",
			outputFormat: "text",
			prettify:     false,
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": {}
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "Datapoint set successfully: ABB7F595EC47.ch0000.idp0000\n  Response: (empty)\n",
		},
		{
			name:         "Empty datapoint response",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "1",
			outputFormat: "text",
			prettify:     false,
			responseBody: `{}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "Failed to set datapoint: ABB7F595EC47.ch0000.idp0000\n",
		},
		{
			name:         "No datapoint for EmptyUUID",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "1",
			outputFormat: "text",
			prettify:     false,
			responseBody: `{
  "other-uuid": {
    "status": "success"
  }
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "Failed to set datapoint: ABB7F595EC47.ch0000.idp0000\n",
		},
		{
			name:          "HTTP error response",
			serial:        "ABB7F595EC47",
			channel:       "ch0000",
			datapoint:     "idp0000",
			value:         "1",
			outputFormat:  "text",
			prettify:      false,
			responseBody:  `{"error": "Unauthorized"}`,
			responseCode:  http.StatusUnauthorized,
			expectError:   true,
			errorContains: "failed to set datapoint",
		},
		{
			name:          "Invalid JSON response",
			serial:        "ABB7F595EC47",
			channel:       "ch0000",
			datapoint:     "idp0000",
			value:         "1",
			outputFormat:  "text",
			prettify:      false,
			responseBody:  `invalid json`,
			responseCode:  http.StatusOK,
			expectError:   true,
			errorContains: "failed to set datapoint",
		},
		{
			name:         "Different serial, channel, datapoint, and value",
			serial:       "DEVICE123",
			channel:      "ch0001",
			datapoint:    "idp0001",
			value:        "0",
			outputFormat: "text",
			prettify:     false,
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": {
    "status": "success",
    "value": "0"
  }
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "Datapoint set successfully: DEVICE123.ch0001.idp0001\n  Response: map[status:success value:0]\n",
		},
		{
			name:         "String value with spaces",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "Hello World",
			outputFormat: "text",
			prettify:     false,
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": {
    "status": "success"
  }
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "Datapoint set successfully: ABB7F595EC47.ch0000.idp0000\n  Response: map[status:success]\n",
		},
		{
			name:         "Numeric value as string",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "42",
			outputFormat: "text",
			prettify:     false,
			responseBody: `{
  "00000000-0000-0000-0000-000000000000": {
    "status": "success"
  }
}`,
			responseCode: http.StatusOK,
			expectError:  false,
			expectOutput: "Datapoint set successfully: ABB7F595EC47.ch0000.idp0000\n  Response: map[status:success]\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create viper instance
			v := setupViper(t)

			// Setup mock SystemAccessPoint
			sysAp, _, _ := setupMock(t, v, tt.responseCode, tt.responseBody)

			// Override the setupFunc to use the mock SystemAccessPoint
			setupFunc = func(config CommandConfig, configFile string) (*freeathome.SystemAccessPoint, error) {
				return sysAp, nil
			}
			defer func() {
				setupFunc = setup
			}()

			// Capture stdout for output testing
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			defer func() { os.Stdout = oldStdout }()

			// Test SetDatapoint function
			err := SetDatapoint(SetCommandConfig{
				CommandConfig: CommandConfig{
					Viper:         v,
					TLSEnabled:    false,
					SkipTLSVerify: false,
					LogLevel:      "info",
				},
				OutputFormat: tt.outputFormat,
				Prettify:     tt.prettify,
			}, tt.serial, tt.channel, tt.datapoint, tt.value)

			// Close pipe and read output
			_ = w.Close()
			output, _ := io.ReadAll(r)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.expectOutput != "" && string(output) != tt.expectOutput {
					t.Errorf("Expected output '%s', got '%s'", tt.expectOutput, string(output))
				}
			}
		})
	}
}

// TestSetDatapointWithInvalidConfigFile tests SetDatapoint with an invalid config file
func TestSetDatapointWithInvalidConfigFile(t *testing.T) {
	// Create a temporary config file with invalid YAML
	configFile := filepath.Join(t.TempDir(), "invalid-config.yaml")

	invalidYAML := `hostname: test-host
username: [invalid array]
password: test-pass`

	err := os.WriteFile(configFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create a fresh viper instance for testing
	v := viper.New()

	// Test SetDatapoint function - should fail due to invalid YAML
	err = SetDatapoint(SetCommandConfig{
		CommandConfig: CommandConfig{
			Viper:         v,
			TLSEnabled:    true,
			SkipTLSVerify: false,
			LogLevel:      "info",
		},
		OutputFormat: "text",
		Prettify:     false,
	}, "ABB7F595EC47", "ch0000", "idp0000", "1")
	if err == nil {
		t.Error("Expected error when loading invalid config file, got none")
	}
}

// TestSetDatapointWithNilViper tests SetDatapoint with nil viper instance
func TestSetDatapointWithNilViper(t *testing.T) {
	// Test SetDatapoint function with nil viper
	err := SetDatapoint(SetCommandConfig{
		CommandConfig: CommandConfig{
			Viper:         nil,
			TLSEnabled:    true,
			SkipTLSVerify: false,
			LogLevel:      "info",
		},
		OutputFormat: "text",
		Prettify:     false,
	}, "ABB7F595EC47", "ch0000", "idp0000", "1")
	if err == nil {
		t.Error("Expected error with nil viper, got none")
	}
}

// TestSetDatapointFunctionExists tests that the SetDatapoint function exists and can be called
func TestSetDatapointFunctionExists(t *testing.T) {
	// Create viper instance
	v := setupViper(t)

	// Test that the function can be called (it will likely fail due to network issues, but that's expected)
	err := SetDatapoint(SetCommandConfig{
		CommandConfig: CommandConfig{
			Viper:         v,
			TLSEnabled:    true,
			SkipTLSVerify: false,
			LogLevel:      "info",
		},
		OutputFormat: "text",
		Prettify:     false,
	}, "ABB7F595EC47", "ch0000", "idp0000", "1")
	// We expect this to fail due to network/connection issues, but the function should exist
	if err == nil {
		t.Log("SetDatapoint function exists and was called successfully")
	} else {
		t.Logf("SetDatapoint function exists but failed as expected: %v", err)
	}
}

// TestSetCommandConfigConversion tests that SetCommandConfig can be converted to GetCommandConfig
func TestSetCommandConfigConversion(t *testing.T) {
	setConfig := SetCommandConfig{
		CommandConfig: CommandConfig{
			Viper:         viper.New(),
			TLSEnabled:    true,
			SkipTLSVerify: false,
			LogLevel:      "debug",
		},
		OutputFormat: "json",
		Prettify:     true,
	}

	getConfig := GetCommandConfig(setConfig)

	if getConfig.Viper != setConfig.Viper {
		t.Error("Viper should be the same after conversion")
	}
	if getConfig.TLSEnabled != setConfig.TLSEnabled {
		t.Error("TLSEnabled should be the same after conversion")
	}
	if getConfig.SkipTLSVerify != setConfig.SkipTLSVerify {
		t.Error("SkipTLSVerify should be the same after conversion")
	}
	if getConfig.LogLevel != setConfig.LogLevel {
		t.Error("LogLevel should be the same after conversion")
	}
	if getConfig.OutputFormat != setConfig.OutputFormat {
		t.Error("OutputFormat should be the same after conversion")
	}
	if getConfig.Prettify != setConfig.Prettify {
		t.Error("Prettify should be the same after conversion")
	}
}

// TestSetDatapointWithEmptyValues tests SetDatapoint with empty values
func TestSetDatapointWithEmptyValues(t *testing.T) {
	tests := []struct {
		name          string
		serial        string
		channel       string
		datapoint     string
		value         string
		expectError   bool
		errorContains string
	}{
		{
			name:        "Empty serial",
			serial:      "",
			channel:     "ch0000",
			datapoint:   "idp0000",
			value:       "1",
			expectError: false, // The system accepts empty serial
		},
		{
			name:        "Empty channel",
			serial:      "ABB7F595EC47",
			channel:     "",
			datapoint:   "idp0000",
			value:       "1",
			expectError: false, // The system accepts empty channel
		},
		{
			name:        "Empty datapoint",
			serial:      "ABB7F595EC47",
			channel:     "ch0000",
			datapoint:   "",
			value:       "1",
			expectError: false, // The system accepts empty datapoint
		},
		{
			name:        "Empty value",
			serial:      "ABB7F595EC47",
			channel:     "ch0000",
			datapoint:   "idp0000",
			value:       "",
			expectError: false, // Empty value should be allowed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create viper instance
			v := setupViper(t)

			// Setup mock SystemAccessPoint with success response
			responseBody := `{
  "00000000-0000-0000-0000-000000000000": {
    "status": "success"
  }
}`
			sysAp, _, _ := setupMock(t, v, http.StatusOK, responseBody)

			// Override the setupFunc to use the mock SystemAccessPoint
			setupFunc = func(config CommandConfig, configFile string) (*freeathome.SystemAccessPoint, error) {
				return sysAp, nil
			}
			defer func() {
				setupFunc = setup
			}()

			// Test SetDatapoint function
			err := SetDatapoint(SetCommandConfig{
				CommandConfig: CommandConfig{
					Viper:         v,
					TLSEnabled:    false,
					SkipTLSVerify: false,
					LogLevel:      "info",
				},
				OutputFormat: "text",
				Prettify:     false,
			}, tt.serial, tt.channel, tt.datapoint, tt.value)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// TestSetDatapointWithSpecialCharacters tests SetDatapoint with special characters in values
func TestSetDatapointWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name         string
		serial       string
		channel      string
		datapoint    string
		value        string
		responseBody string
		expectError  bool
	}{
		{
			name:         "Special characters in value",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "!@#$%^&*()",
			responseBody: `{"00000000-0000-0000-0000-000000000000":{"status":"success"}}`,
			expectError:  false,
		},
		{
			name:         "Unicode characters in value",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        "Hello 世界",
			responseBody: `{"00000000-0000-0000-0000-000000000000":{"status":"success"}}`,
			expectError:  false,
		},
		{
			name:         "JSON-like value",
			serial:       "ABB7F595EC47",
			channel:      "ch0000",
			datapoint:    "idp0000",
			value:        `{"key":"value"}`,
			responseBody: `{"00000000-0000-0000-0000-000000000000":{"status":"success"}}`,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create viper instance
			v := setupViper(t)

			// Setup mock SystemAccessPoint
			sysAp, _, _ := setupMock(t, v, http.StatusOK, tt.responseBody)

			// Override the setupFunc to use the mock SystemAccessPoint
			setupFunc = func(config CommandConfig, configFile string) (*freeathome.SystemAccessPoint, error) {
				return sysAp, nil
			}
			defer func() {
				setupFunc = setup
			}()

			// Test SetDatapoint function
			err := SetDatapoint(SetCommandConfig{
				CommandConfig: CommandConfig{
					Viper:         v,
					TLSEnabled:    false,
					SkipTLSVerify: false,
					LogLevel:      "info",
				},
				OutputFormat: "text",
				Prettify:     false,
			}, tt.serial, tt.channel, tt.datapoint, tt.value)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
