package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lion-and-bear/freeathome/v2/internal/cli"
)

var (
	setCmd = &cobra.Command{
		Use:   "set",
		Short: "Set data on the free@home system access point",
		Long:  `Set various types of data on the free@home system access point.`,
	}

	datapointSetCmd = &cobra.Command{
		Use:     "datapoint [serial] [channel] [datapoint] [value]",
		Aliases: []string{"dp"},
		Short:   "Set a specific datapoint value on the system access point",
		Long:    `Set the value of a specific datapoint by its serial number, channel, datapoint identifier, and value.`,
		Args:    cobra.ExactArgs(4),
		RunE:    runSetDatapoint,
	}
)

func init() {
	rootCmd.AddCommand(setCmd)

	// Add subcommands
	setCmd.AddCommand(datapointSetCmd)

	// Add TLS configuration flags
	setCmd.PersistentFlags().BoolVar(&tlsEnabled, "tls", true, "Enable TLS for connection")
	setCmd.PersistentFlags().BoolVar(&skipTLSVerify, "skip-tls-verify", false, "Skip TLS certificate verification")

	// Add logging configuration flag
	setCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level (debug, info, warn, error)")

	// Add output format flag
	setCmd.PersistentFlags().StringVar(&outputFormat, "output", "json", "Set the output format (json, text)")

	// Add prettify flag
	setCmd.PersistentFlags().BoolVar(&prettify, "prettify", false, "Prettify JSON output with indentation. Only used for JSON output.")
}

func runSetDatapoint(cmd *cobra.Command, args []string) error {
	return cli.SetDatapoint(cli.SetCommandConfig{
		CommandConfig: cli.CommandConfig{
			Viper:         viper.GetViper(),
			TLSEnabled:    tlsEnabled,
			SkipTLSVerify: skipTLSVerify,
			LogLevel:      logLevel,
		},
		OutputFormat: outputFormat,
		Prettify:     prettify,
	}, args[0], args[1], args[2], args[3])
}
