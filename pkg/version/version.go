package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version, set via ldflags at build time.
	Version = "dev"
	// Commit is the git commit hash, set via ldflags at build time.
	Commit = "unknown"
	// BuildDate is the UTC build timestamp, set via ldflags at build time.
	BuildDate = "unknown"
)

// Info returns a formatted version string.
func Info() string {
	return fmt.Sprintf("pain_tz %s (commit: %s, built: %s, go: %s)",
		Version, Commit, BuildDate, runtime.Version())
}
