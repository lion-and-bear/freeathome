package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/lion-and-bear/freeathome/v2/internal/cli"
)

var (
	// Configuration file path
	cfgFile string

	// Configuration values
	hostname string
	username string
	password string

	// Non-interactive mode flag
	nonInteractive bool

	configureCmd = &cobra.Command{
		Use:   "configure",
		Short: "Configure connection settings for free@home system access point",
		Long: `Configure the connection settings for your free@home system access point.
This command allows you to set the hostname, username, and password
using flags, environment variables, or a YAML configuration file.

Examples:
  free@home configure --hostname 192.168.1.100 --username admin --password mypass
  free@home configure --config ~/.freeathome/config.yaml
  free@home configure --non-interactive --hostname 192.168.1.100 --username admin --password mypass
  export FREEATHOME_HOSTNAME=192.168.1.100
  export FREEATHOME_USERNAME=admin
  export FREEATHOME_PASSWORD=mypass
  free@home configure`,
		RunE: runConfigure,
	}

	showCmd = &cobra.Command{
		Use:   "show",
		Short: "Show current configuration settings",
		Long:  `Display the current configuration settings for the free@home system access point.`,
		RunE:  runShow,
	}
)

func init() {
	rootCmd.AddCommand(configureCmd)

	// Add subcommands
	configureCmd.AddCommand(showCmd)

	// Add flags
	configureCmd.Flags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.freeathome/config.yaml)")
	configureCmd.Flags().StringVar(&hostname, "hostname", "", "free@home system hostname or IP address")
	configureCmd.Flags().StringVar(&username, "username", "", "username for authentication")
	configureCmd.Flags().StringVar(&password, "password", "", "password for authentication")
	configureCmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "fail instead of prompting for missing configuration values")

	// Bind flags to viper
	_ = viper.BindPFlag("hostname", configureCmd.Flags().Lookup("hostname"))
	_ = viper.BindPFlag("username", configureCmd.Flags().Lookup("username"))
	_ = viper.BindPFlag("password", configureCmd.Flags().Lookup("password"))
}

func runConfigure(cmd *cobra.Command, args []string) error {
	return cli.Configure(viper.GetViper(), cfgFile, hostname, username, password, nonInteractive)
}

func runShow(cmd *cobra.Command, args []string) error {
	return cli.ShowConfiguration(viper.GetViper(), cfgFile)
}
