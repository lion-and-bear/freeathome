package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/lion-and-bear/freeathome/v2/pkg/freeathome"
	"github.com/lion-and-bear/freeathome/v2/pkg/models"
	"golang.org/x/term"
)

var setupFunc = setup

// GetCommandConfig is a struct that contains the configuration for the get command
type GetCommandConfig struct {
	CommandConfig
	OutputFormat string
	Prettify     bool
}

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		// Default to info level for unknown values
		return slog.LevelInfo
	}
}

// handleSysApError provides consistent error handling for system access point operations
func handleSysApError(err error, operation string, tlsEnabled, skipTLSVerify bool) error {
	if err == nil {
		return nil
	}

	// Provide helpful error message for TLS issues
	if tlsEnabled && !skipTLSVerify {
		return fmt.Errorf("failed to %s: %w\n\nIf you're getting TLS certificate errors, try:\n  - Using --skip-tls-verify flag\n  - Using --tls=false to use HTTP instead of HTTPS", operation, err)
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// outputJSON provides consistent JSON output formatting for system access point operations
func outputJSON(data any, dataType string, prettify bool) error {
	var jsonData []byte
	var err error

	if prettify {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	} else {
		jsonData, err = json.Marshal(data)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal %s to JSON: %w", dataType, err)
	}
	fmt.Println(string(jsonData))
	return nil
}

func setup(config CommandConfig, configFile string) (*freeathome.SystemAccessPoint, error) {
	// Load configuration
	cfg, err := load(config.Viper, configFile)
	if err != nil {
		return nil, err
	}

	// Check if configuration is complete
	if cfg.Hostname == "" {
		return nil, fmt.Errorf("hostname not configured. Run '%s configure' first", cfg.Executable)
	}
	if cfg.Username == "" {
		return nil, fmt.Errorf("username not configured. Run '%s configure' first", cfg.Executable)
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("password not configured. Run '%s configure' first", cfg.Executable)
	}

	// Create a new logger with the specified options
	// Use a colorized handler if the terminal supports colors
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		color.NoColor = true
	}
	handler := freeathome.NewColorHandler(os.Stderr, &slog.HandlerOptions{
		Level: parseLogLevel(config.LogLevel),
	})
	logger := freeathome.NewDefaultLogger(handler)

	// Create system access point client
	sysApConfig := freeathome.NewConfig(cfg.Hostname, cfg.Username, cfg.Password)
	sysApConfig.TLSEnabled = config.TLSEnabled
	sysApConfig.SkipTLSVerify = config.SkipTLSVerify
	sysApConfig.Logger = logger
	return freeathome.NewSystemAccessPoint(sysApConfig)
}

// GetDeviceList retrieves and displays the device list
func GetDeviceList(config GetCommandConfig) error {
	// Setup system access point
	sysAp, err := setupFunc(config.CommandConfig, "")
	if err != nil {
		return err
	}

	// Get device list
	deviceList, err := sysAp.GetDeviceList()
	if err != nil {
		return handleSysApError(err, "get device list", config.TLSEnabled, config.SkipTLSVerify)
	}

	// Output depending on output format
	if config.OutputFormat == "json" {
		return outputJSON(deviceList, "device list", config.Prettify)
	}

	// Check if device list is empty
	if deviceList == nil || len(*deviceList) == 0 {
		fmt.Println("No devices found")
		return nil
	}

	// Get devices for the system access point (using EmptyUUID as key)
	devices, exists := (*deviceList)[models.EmptyUUID]
	if !exists {
		fmt.Println("No devices found for system access point")
		return nil
	}

	if len(devices) == 0 {
		fmt.Println("No devices found")
		return nil
	}

	// Output as plain text (one device per line)
	for _, deviceSerial := range devices {
		fmt.Println(deviceSerial)
	}

	return nil
}

// GetConfiguration retrieves and displays the configuration
func GetConfiguration(config GetCommandConfig) error {
	// Setup system access point
	sysAp, err := setupFunc(config.CommandConfig, "")
	if err != nil {
		return err
	}

	// Get configuration
	configuration, err := sysAp.GetConfiguration()
	if err != nil {
		return handleSysApError(err, "get configuration", config.TLSEnabled, config.SkipTLSVerify)
	}

	// Output depending on output format
	if config.OutputFormat == "json" {
		return outputJSON(configuration, "configuration", config.Prettify)
	}

	// Check if configuration is empty
	if configuration == nil || len(*configuration) == 0 {
		fmt.Println("No configuration found")
		return nil
	}

	// Output as plain text (one system access point per line)
	for sysApID, sysAp := range *configuration {
		fmt.Printf("System Access Point ID: %s\n", sysApID)
		fmt.Printf("  Name: %s\n", sysAp.SysApName)
		fmt.Printf("  Devices: %d\n", len(sysAp.Devices))
		fmt.Printf("  Users: %d\n", len(sysAp.Users))
		fmt.Printf("  Floors: %d\n", len(sysAp.Floorplan.Floors))
		fmt.Println()
	}

	return nil
}

// GetDevice retrieves and displays a specific device by serial number
func GetDevice(config GetCommandConfig, serial string) error {
	// Setup system access point
	sysAp, err := setupFunc(config.CommandConfig, "")
	if err != nil {
		return err
	}

	// Get device
	device, err := sysAp.GetDevice(serial)
	if err != nil {
		return handleSysApError(err, "get device", config.TLSEnabled, config.SkipTLSVerify)
	}

	// Output depending on output format
	if config.OutputFormat == "json" {
		return outputJSON(device, "device", config.Prettify)
	}

	// Check if device is empty
	var deviceNotFound = fmt.Sprintf("No device found with serial: %s", serial)
	if device == nil || len(*device) == 0 {
		fmt.Println(deviceNotFound)
		return nil
	}

	// Get devices for the system access point (using EmptyUUID as key)
	devices, exists := (*device)[models.EmptyUUID]
	if !exists {
		fmt.Println(deviceNotFound)
		return nil
	}

	// Check if the specific device exists
	deviceData, deviceExists := devices.Devices[serial]
	if !deviceExists {
		fmt.Println(deviceNotFound)
		return nil
	}

	// Output as plain text
	fmt.Printf("Device Serial: %s\n", serial)
	if deviceData.DisplayName != nil {
		fmt.Printf("  Display Name: %s\n", *deviceData.DisplayName)
	}
	if deviceData.Room != nil {
		fmt.Printf("  Room: %s\n", *deviceData.Room)
	}
	if deviceData.Floor != nil {
		fmt.Printf("  Floor: %s\n", *deviceData.Floor)
	}
	if deviceData.Interface != nil {
		fmt.Printf("  Interface: %s\n", *deviceData.Interface)
	}
	if deviceData.NativeID != nil {
		fmt.Printf("  Native ID: %s\n", *deviceData.NativeID)
	}
	if deviceData.Channels != nil {
		fmt.Printf("  Channels: %d\n", len(*deviceData.Channels))
	}
	if deviceData.Parameters != nil {
		fmt.Printf("  Parameters: %d\n", len(*deviceData.Parameters))
	}

	return nil
}

// GetDatapoint retrieves and displays a specific datapoint
func GetDatapoint(config GetCommandConfig, serial string, channel string, datapoint string) error {
	// Setup system access point
	sysAp, err := setupFunc(config.CommandConfig, "")
	if err != nil {
		return err
	}

	// Get datapoint
	datapointResponse, err := sysAp.GetDatapoint(serial, channel, datapoint)
	if err != nil {
		return handleSysApError(err, "get datapoint", config.TLSEnabled, config.SkipTLSVerify)
	}

	// Output depending on output format
	if config.OutputFormat == "json" {
		return outputJSON(datapointResponse, "datapoint", config.Prettify)
	}

	// Check if datapoint response is empty
	if datapointResponse == nil || len(*datapointResponse) == 0 {
		fmt.Printf("No datapoint found: %s.%s.%s\n", serial, channel, datapoint)
		return nil
	}

	// Get datapoint for the system access point (using EmptyUUID as key)
	datapointData, exists := (*datapointResponse)[models.EmptyUUID]
	if !exists {
		fmt.Printf("No datapoint found: %s.%s.%s\n", serial, channel, datapoint)
		return nil
	}

	// Output as plain text
	fmt.Printf("Datapoint: %s.%s.%s\n", serial, channel, datapoint)
	if len(datapointData.Values) > 0 {
		fmt.Printf("  Values: %v\n", datapointData.Values)
	} else {
		fmt.Printf("  Values: (empty)\n")
	}

	return nil
}
