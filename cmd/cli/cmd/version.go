package cmd

import (
	"fmt"

	internal "github.com/lion-and-bear/freeathome/v2/internal"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Long:  `Print the version of the free@home CLI`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("free@home CLI v%s-%s\n", internal.Version, internal.Commit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
