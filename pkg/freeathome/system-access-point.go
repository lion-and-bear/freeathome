package freeathome

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/go-resty/resty/v2"

	"github.com/lion-and-bear/freeathome/v2/pkg/models"
)

// Config represents the configuration for a SystemAccessPoint
type Config struct {
	// Hostname is the hostname or IP address of the system access point
	Hostname string
	// Username is the username for authentication
	Username string
	// Password is the password for authentication
	Password string
	// TLSEnabled indicates whether TLS is enabled for communication
	TLSEnabled bool
	// SkipTLSVerify indicates whether TLS certificate verification should be skipped
	SkipTLSVerify bool
	// VerboseErrors indicates whether verbose errors should be logged
	VerboseErrors bool
	// Logger is the logger to use for logging messages
	Logger models.Logger
	// Client is the REST client to use (optional, will create default if nil)
	Client *resty.Client
}

// NewConfig creates a new Config with default values
func NewConfig(hostname, username, password string) *Config {
	return &Config{
		Hostname:      hostname,
		Username:      username,
		Password:      password,
		TLSEnabled:    true,
		SkipTLSVerify: false,
		VerboseErrors: false,
		Logger:        nil,
		Client:        nil,
	}
}

// SystemAccessPoint represents a system access point that can be used to communicate with a free@home system.
type SystemAccessPoint struct {
	UUID string
	// config contains the configuration for the system access point
	config *Config
	// datapointRegex is the regular expression that is used to match datapoint keys.
	datapointRegex *regexp.Regexp
	// clock provides time operations that can be mocked in tests
	clock clock
	// onError is a callback function that is called when an error occurs.
	onError func(error)
}

// NewSystemAccessPoint creates a new SystemAccessPoint with the specified configuration.
func NewSystemAccessPoint(config *Config) (*SystemAccessPoint, error) {
	if config == nil {
		return nil, errors.New("config cannot be nil")
	}

	// Set default logger if not provided
	if config.Logger == nil {
		config.Logger = NewDefaultLogger(nil)
		config.Logger.Warn("No logger provided for SystemAccessPoint. Using default logger.")
	}

	// Create REST client with basic auth
	if config.Client == nil {
		config.Client = resty.New()
	}
	config.Client.SetBasicAuth(config.Username, config.Password)

	// Configure TLS settings if TLS is enabled
	if config.TLSEnabled && config.SkipTLSVerify {
		config.Logger.Warn("TLS is enabled but certificate verification is disabled, this is not recommended!")
		config.Client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return &SystemAccessPoint{
		UUID:           models.EmptyUUID,
		config:         config,
		datapointRegex: regexp.MustCompile(models.DatapointPattern),
		clock:          &realClock{},
	}, nil
}

// MustNewSystemAccessPoint creates a new SystemAccessPoint with the specified configuration.
// It panics if an error occurs.
func MustNewSystemAccessPoint(config *Config) *SystemAccessPoint {
	sysap, err := NewSystemAccessPoint(config)
	// The error can only occur if the config is nil, which is considered a programming error.
	// If you are not sure if the config is nil, use NewSystemAccessPoint instead.
	if err != nil {
		panic(err)
	}
	return sysap
}

// NewSystemAccessPointWithDefaults creates a new SystemAccessPoint with minimal configuration
func NewSystemAccessPointWithDefaults(hostname, username, password string) *SystemAccessPoint {
	config := NewConfig(hostname, username, password)
	return MustNewSystemAccessPoint(config)
}

// emitError is a helper function to emit errors using the onError callback.
func (sysAp *SystemAccessPoint) emitError(err error) {
	if sysAp.onError != nil {
		sysAp.onError(err)
	}
}

// HostName returns the host name of the system access point.
func (sysAp *SystemAccessPoint) GetHostName() string {
	return sysAp.config.Hostname
}

// TlsEnabled returns whether TLS is enabled for communication with the system access point.
func (sysAp *SystemAccessPoint) GetTlsEnabled() bool {
	return sysAp.config.TLSEnabled
}

// SkipTLSVerify returns whether TLS certificate verification should be skipped.
func (sysAp *SystemAccessPoint) GetSkipTLSVerify() bool {
	return sysAp.config.SkipTLSVerify
}

// VerboseErrors returns whether verbose errors should be logged.
func (sysAp *SystemAccessPoint) GetVerboseErrors() bool {
	return sysAp.config.VerboseErrors
}

// GetUrl constructs a URL string for the SystemAccessPoint based on the provided path.
// It uses the appropriate protocol (http or https) depending on whether TLS is enabled.
//
// Parameters:
//   - path: The specific API endpoint path to be appended to the base URL.
//
// Returns:
//   - A formatted URL string that includes the protocol, hostname, and the provided path.
func (sysAp *SystemAccessPoint) GetUrl(path string) string {
	var protocol string
	if sysAp.config.TLSEnabled {
		protocol = "https"
	} else {
		protocol = "http"
	}

	return fmt.Sprintf("%s://%s/fhapi/v1/api/rest/%s", protocol, sysAp.config.Hostname, path)
}

// CreateVirtualDevice creates a new virtual device on the System Access Point (SysAP) with the specified serial number.
// It sends a PUT request containing the provided VirtualDevice data to the SysAP API.
// On success, it returns the response containing details of the created virtual device.
// If an error occurs during the request or response parsing, it logs the error, emits an error event, and returns the error.
//
// Parameters:
//   - serial: The serial number of the virtual device to be created.
//   - virtualDevice: A pointer to the VirtualDevice struct containing the device configuration.
//
// Returns:
//   - *models.VirtualDeviceResponse: Pointer to the response struct with details of the created virtual device.
//   - error: An error object if the operation fails, otherwise nil.
func (sysAp *SystemAccessPoint) CreateVirtualDevice(serial string, virtualDevice *models.VirtualDevice) (*models.VirtualDeviceResponse, error) {
	resp, err := sysAp.config.Client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "serial": serial}).
		SetBody(virtualDevice).
		Put(sysAp.GetUrl("virtualdevice/{uuid}/{serial}"))

	return deserializeRestResponse[models.VirtualDeviceResponse](sysAp, resp, err, "failed to create virtual device")
}

// GetConfiguration retrieves the configuration from the system access point.
// It sends a GET request to the "configuration" endpoint and unmarshals the response
// into a models.Configuration object.
//
// Returns a pointer to the Configuration object and an error if any occurred during the request.
//
// Possible errors include network issues, non-2xx HTTP responses, or unmarshalling errors.
func (sysAp *SystemAccessPoint) GetConfiguration() (*models.Configuration, error) {
	resp, err := sysAp.config.Client.R().Get(sysAp.GetUrl("configuration"))

	return deserializeRestResponse[models.Configuration](sysAp, resp, err, "failed to get configuration")
}

// GetDeviceList retrieves the list of devices from the system access point.
// It sends a GET request to the "devicelist" endpoint and unmarshals the response
// into a DeviceList model.
//
// Returns:
//   - *models.DeviceList: A pointer to the DeviceList model containing the list of devices.
//   - error: An error if the request fails or the response contains an error.
func (sysAp *SystemAccessPoint) GetDeviceList() (*models.DeviceList, error) {
	resp, err := sysAp.config.Client.R().Get(sysAp.GetUrl("devicelist"))

	return deserializeRestResponse[models.DeviceList](sysAp, resp, err, "failed to get device list")
}

// GetDevice retrieves a device with the specified serial number from the system access point.
// It sends a GET request to the appropriate endpoint and parses the response into a DeviceResponse model.
// Returns a pointer to the DeviceResponse and an error if the request fails or the response cannot be parsed.
func (sysAp *SystemAccessPoint) GetDevice(serial string) (*models.DeviceResponse, error) {
	resp, err := sysAp.config.Client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "serial": serial}).
		Get(sysAp.GetUrl("device/{uuid}/{serial}"))

	return deserializeRestResponse[models.DeviceResponse](sysAp, resp, err, "failed to get device")
}

// GetDatapoint retrieves a datapoint from the System Access Point using the provided serial number, channel, and datapoint identifiers.
// It sends a GET request to the corresponding endpoint and parses the response into a models.Datapoint object.
// If the request fails or the response cannot be parsed, an error is returned and logged.
//
// Parameters:
//
//	serial   - The serial number of the device.
//	channel  - The channel identifier.
//	datapoint - The datapoint identifier.
//
// Returns:
//
//	*models.Datapoint - The retrieved datapoint object.
//	error             - An error if the request or parsing fails.
func (sysAp *SystemAccessPoint) GetDatapoint(serial string, channel string, datapoint string) (*models.GetDataPointResponse, error) {
	resp, err := sysAp.config.Client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "serial": serial, "channel": channel, "datapoint": datapoint}).
		Get(sysAp.GetUrl("datapoint/{uuid}/{serial}.{channel}.{datapoint}"))

	return deserializeRestResponse[models.GetDataPointResponse](sysAp, resp, err, "failed to get datapoint")
}

// SetDatapoint sets the value of a specified datapoint for a given device channel.
// It sends a PUT request to the System Access Point (SysAP) to update the datapoint value.
//
// Parameters:
//
//	serial    - The serial number of the target device.
//	channel   - The channel identifier of the device.
//	datapoint - The datapoint identifier to be set.
//	value     - The value to set for the datapoint.
//
// Returns:
//
//	*models.SetDataPointResponse - The response from the SysAP after setting the datapoint.
//	error                        - An error if the request fails or the response cannot be parsed.
func (sysAp *SystemAccessPoint) SetDatapoint(serial string, channel string, datapoint string, value string) (*models.SetDataPointResponse, error) {
	resp, err := sysAp.config.Client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "serial": serial, "channel": channel, "datapoint": datapoint}).
		SetBody(value).
		Put(sysAp.GetUrl("datapoint/{uuid}/{serial}.{channel}.{datapoint}"))

	return deserializeRestResponse[models.SetDataPointResponse](sysAp, resp, err, "failed to set datapoint")
}

// TriggerProxyDevice sends a request to trigger an action on a proxy device identified by its class and serial number.
// It constructs the request URL using the SystemAccessPoint's UUID, the device class, serial, and the specified action.
// The method returns the parsed DeviceResponse on success, or an error if the request fails or the response cannot be parsed.
//
// Parameters:
//   - class:  The class of the proxy device.
//   - serial: The serial number of the proxy device.
//   - action: The action to trigger on the proxy device.
//
// Returns:
//   - *models.DeviceResponse: The response from the device if the action is successful.
//   - error: An error if the request fails or the response cannot be parsed.
func (sysAp *SystemAccessPoint) TriggerProxyDevice(class string, serial string, action string) (*models.DeviceResponse, error) {
	resp, err := sysAp.config.Client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "class": class, "serial": serial, "action": action}).
		Get(sysAp.GetUrl("proxydevice/{uuid}/{class}/{serial}/action/{action}"))

	return deserializeRestResponse[models.DeviceResponse](sysAp, resp, err, "failed to trigger proxy device")
}

// SetProxyDeviceValue sets the value of a proxy device identified by its class and serial number.
// It sends a PUT request to the system access point's API and returns the device response.
//
// Parameters:
//   - class:  The device class identifier.
//   - serial: The serial number of the device.
//   - value:  The value to set for the device.
//
// Returns:
//   - *models.DeviceResponse: The response from the device if the operation is successful.
//   - error: An error if the request fails or the response cannot be parsed.
func (sysAp *SystemAccessPoint) SetProxyDeviceValue(class string, serial string, value string) (*models.DeviceResponse, error) {
	resp, err := sysAp.config.Client.R().
		SetPathParams(map[string]string{"uuid": sysAp.UUID, "class": class, "serial": serial, "value": value}).
		Put(sysAp.GetUrl("proxydevice/{uuid}/{class}/{serial}/value/{value}"))

	return deserializeRestResponse[models.DeviceResponse](sysAp, resp, err, "failed to set proxy device value")
}

func deserializeRestResponse[T any](sysAp *SystemAccessPoint, resp *resty.Response, err error, errorMessage string) (*T, error) {
	// Check for errors
	if err != nil {
		sysAp.config.Logger.Error(errorMessage, "error", err)
		sysAp.emitError(err)
		return nil, err
	}

	if resp.IsError() {
		sysAp.config.Logger.Error(errorMessage, "status", resp.Status(), "body", resp.String())
		return nil, fmt.Errorf("%s: %s", errorMessage, resp.String())
	}

	var object T
	if err := json.Unmarshal(resp.Body(), &object); err != nil {
		sysAp.config.Logger.Error("failed to parse response body", "error", err)
		sysAp.emitError(err)
		return nil, err
	}

	return &object, nil
}
