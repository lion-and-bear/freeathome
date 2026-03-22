package cli

import (
	"fmt"

	"github.com/lion-and-bear/freeathome/v2/pkg/models"
)

// SetCommandConfig is a struct that contains the configuration for the set command
type SetCommandConfig struct {
	CommandConfig
	OutputFormat string
	Prettify     bool
}

// SetDatapoint sets a specific datapoint value
func SetDatapoint(config SetCommandConfig, serial string, channel string, datapoint string, value string) error {
	// Setup system access point
	sysAp, err := setupFunc(config.CommandConfig, "")
	if err != nil {
		return err
	}

	// Set datapoint
	datapointResponse, err := sysAp.SetDatapoint(serial, channel, datapoint, value)
	if err != nil {
		return handleSysApError(err, "set datapoint", config.TLSEnabled, config.SkipTLSVerify)
	}

	// Output depending on output format
	if config.OutputFormat == "json" {
		return outputJSON(datapointResponse, "datapoint", config.Prettify)
	}

	// Check if datapoint response is empty
	if datapointResponse == nil || len(*datapointResponse) == 0 {
		fmt.Printf("Failed to set datapoint: %s.%s.%s\n", serial, channel, datapoint)
		return nil
	}

	// Get datapoint for the system access point (using EmptyUUID as key)
	datapointData, exists := (*datapointResponse)[models.EmptyUUID]
	if !exists {
		fmt.Printf("Failed to set datapoint: %s.%s.%s\n", serial, channel, datapoint)
		return nil
	}

	// Output as plain text
	fmt.Printf("Datapoint set successfully: %s.%s.%s\n", serial, channel, datapoint)
	if len(datapointData) > 0 {
		fmt.Printf("  Response: %v\n", datapointData)
	} else {
		fmt.Printf("  Response: (empty)\n")
	}

	return nil
}
