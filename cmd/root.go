package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var rootCmd = &cobra.Command{
	Use:   "wrtd",
	Short: "Router statistics collector",
	Long:  `Collection of statistics for prometheus.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
