package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

//nolint:gochecknoglobals
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Monitoring application",

	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("config")

		fmt.Println("rest called", path) //nolint:forbidigo
	},
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(restCmd)

	restCmd.Flags().
		String(
			"config", "/etc/wrtd/config.yaml", `Service configuration file`)
}
