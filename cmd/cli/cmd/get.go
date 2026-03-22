package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lion-and-bear/freeathome/v2/internal/cli"
)

var (
	// TLS configuration flags
	tlsEnabled    bool
	skipTLSVerify bool
	// Logging configuration
	logLevel string
	// Output format configuration
	outputFormat string
	// JSON output configuration
	prettify bool

	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Get data from the free@home system access point",
		Long:  `Retrieve various types of data from the free@home system access point.`,
	}

	devicelistCmd = &cobra.Command{
		Use:     "devicelist",
		Aliases: []string{"dl"},
		Short:   "Get the list of devices from the system access point",
		Long:    `Retrieve and display the list of devices from the free@home system access point.`,
		RunE:    runGetDeviceList,
	}

	configurationCmd = &cobra.Command{
		Use:     "configuration",
		Aliases: []string{"config", "cfg"},
		Short:   "Get the configuration from the system access point",
		Long:    `Retrieve and display the configuration from the free@home system access point.`,
		RunE:    runGetConfiguration,
	}

	deviceCmd = &cobra.Command{
		Use:     "device [serial]",
		Aliases: []string{"dev"},
		Short:   "Get a specific device from the system access point",
		Long:    `Retrieve and display information about a specific device by its serial number.`,
		Args:    cobra.ExactArgs(1),
		RunE:    runGetDevice,
	}

	datapointCmd = &cobra.Command{
		Use:     "datapoint [serial] [channel] [datapoint]",
		Aliases: []string{"dp"},
		Short:   "Get a specific datapoint from the system access point",
		Long:    `Retrieve and display information about a specific datapoint by its serial number, channel, and datapoint identifier.`,
		Args:    cobra.ExactArgs(3),
		RunE:    runGetDatapoint,
	}
)

func init() {
	rootCmd.AddCommand(getCmd)

	// Add subcommands
	getCmd.AddCommand(devicelistCmd)
	getCmd.AddCommand(configurationCmd)
	getCmd.AddCommand(deviceCmd)
	getCmd.AddCommand(datapointCmd)

	// Add TLS configuration flags
	getCmd.PersistentFlags().BoolVar(&tlsEnabled, "tls", true, "Enable TLS for connection")
	getCmd.PersistentFlags().BoolVar(&skipTLSVerify, "skip-tls-verify", false, "Skip TLS certificate verification")

	// Add logging configuration flag
	getCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level (debug, info, warn, error)")

	// Add output format flag
	getCmd.PersistentFlags().StringVar(&outputFormat, "output", "json", "Set the output format (json, text)")

	// Add prettify flag
	getCmd.PersistentFlags().BoolVar(&prettify, "prettify", false, "Prettify JSON output with indentation. Only used for JSON output.")
}

func runGetDeviceList(cmd *cobra.Command, args []string) error {
	return cli.GetDeviceList(cli.GetCommandConfig{
		CommandConfig: cli.CommandConfig{
			Viper:         viper.GetViper(),
			TLSEnabled:    tlsEnabled,
			SkipTLSVerify: skipTLSVerify,
			LogLevel:      logLevel,
		},
		OutputFormat: outputFormat,
		Prettify:     prettify,
	})
}

func runGetConfiguration(cmd *cobra.Command, args []string) error {
	return cli.GetConfiguration(cli.GetCommandConfig{
		CommandConfig: cli.CommandConfig{
			Viper:         viper.GetViper(),
			TLSEnabled:    tlsEnabled,
			SkipTLSVerify: skipTLSVerify,
			LogLevel:      logLevel,
		},
		OutputFormat: outputFormat,
		Prettify:     prettify,
	})
}

func runGetDevice(cmd *cobra.Command, args []string) error {
	return cli.GetDevice(cli.GetCommandConfig{
		CommandConfig: cli.CommandConfig{
			Viper:         viper.GetViper(),
			TLSEnabled:    tlsEnabled,
			SkipTLSVerify: skipTLSVerify,
			LogLevel:      logLevel,
		},
		OutputFormat: outputFormat,
		Prettify:     prettify,
	}, args[0])
}

func runGetDatapoint(cmd *cobra.Command, args []string) error {
	return cli.GetDatapoint(cli.GetCommandConfig{
		CommandConfig: cli.CommandConfig{
			Viper:         viper.GetViper(),
			TLSEnabled:    tlsEnabled,
			SkipTLSVerify: skipTLSVerify,
			LogLevel:      logLevel,
		},
		OutputFormat: outputFormat,
		Prettify:     prettify,
	}, args[0], args[1], args[2])
}
