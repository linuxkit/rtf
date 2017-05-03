package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	Version   string
	GitCommit string
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s\n", filepath.Base(os.Args[0]), Version)
		fmt.Printf("commit: %s\n", GitCommit)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
