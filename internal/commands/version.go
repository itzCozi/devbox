package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	Version = "1.0"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Display the version and build information for devbox.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("devbox (v%s)\n", Version)
	},
}

func init() {

}
