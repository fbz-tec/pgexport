package main

import (
	"fmt"
	"runtime"
)

// Version information (set via ldflags during build)
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// GetVersionInfo returns formatted version information
func GetVersionInfo() string {
	return fmt.Sprintf(
		`pgxport version %s
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
