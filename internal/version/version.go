package version

import (
	"fmt"
	"runtime"
)

// Version information (set via ldflags during build)
var (
	AppVersion = "dev"
	BuildTime  = "unknown"
	GitCommit  = "unknown"
)

// GetVersionInfo returns formatted version information
func GetInfo() string {
	return fmt.Sprintf(
		`Version: %s
Build time: %s
Git commit: %s
Go version: %s
OS/Arch: %s/%s`,
		AppVersion,
		BuildTime,
		GitCommit,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}
