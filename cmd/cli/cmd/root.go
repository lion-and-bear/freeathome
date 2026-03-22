package cmd

import (
	"github.com/lion-and-bear/freeathome/v2/internal/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   cli.MustExecutableName(),
	Short: "Interact with ABB free@home devices using the local API",
	Long:  `A CLI tool to interact with ABB free@home devices using the local API.`,
}

func Execute() error {
	return rootCmd.Execute()
}
