package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lion-and-bear/freeathome/v2/internal/cli"
)

var (
	// Monitor-specific flags
	timeout                 int
	maxReconnectionAttempts int
	exponentialBackoff      bool
	monitorRaw              bool
	// Inherit common flags from other commands
	monitorTLSEnabled    bool
	monitorSkipTLSVerify bool
	monitorLogLevel      string
)

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor the free@home system access point via WebSocket",
	Long: `Connect to the free@home system access point via WebSocket and monitor real-time events.
With --raw, each WebSocket text frame is written unchanged to stdout (one line per message); hints and logs go to stderr.`,
	RunE: runMonitor,
}

func init() {
	rootCmd.AddCommand(monitorCmd)

	// Add monitor-specific flags
	monitorCmd.Flags().IntVar(&timeout, "timeout", 30, "WebSocket connection timeout in seconds")
	monitorCmd.Flags().IntVar(&maxReconnectionAttempts, "max-reconnection-attempts", 3, "Maximum number of reconnection attempts before giving up")
	monitorCmd.Flags().BoolVar(&exponentialBackoff, "exponential-backoff", true, "Enable exponential backoff between reconnection attempts")
	monitorCmd.Flags().BoolVar(&monitorRaw, "raw", false, "Stream raw WebSocket text frames to stdout (newline-terminated); use stderr for hints and logs")

	// Add TLS configuration flags
	monitorCmd.Flags().BoolVar(&monitorTLSEnabled, "tls", true, "Enable TLS for connection")
	monitorCmd.Flags().BoolVar(&monitorSkipTLSVerify, "skip-tls-verify", false, "Skip TLS certificate verification")

	// Add logging configuration flag
	monitorCmd.Flags().StringVar(&monitorLogLevel, "log-level", "info", "Set the log level (debug, info, warn, error)")
}

func runMonitor(cmd *cobra.Command, args []string) error {
	return cli.Monitor(cli.MonitorCommandConfig{
		CommandConfig: cli.CommandConfig{
			Viper:         viper.GetViper(),
			TLSEnabled:    monitorTLSEnabled,
			SkipTLSVerify: monitorSkipTLSVerify,
			LogLevel:      monitorLogLevel,
		},
		Timeout:                 timeout,
		MaxReconnectionAttempts: maxReconnectionAttempts,
		ExponentialBackoff:      exponentialBackoff,
		Raw:                     monitorRaw,
	})
}
