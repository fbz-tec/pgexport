package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(GetVersionInfo())
	},
}

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return fmt.Sprintf(
		`Version: %s
Build time: %s
Git commit: %s
Go version: %s
OS/Arch: %s/%s`,
		Version,
		BuildTime,
		GitCommit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}
