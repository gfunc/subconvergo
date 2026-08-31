// Package version holds the build-time version metadata for subconvergo.
// The Version variable is overridden by the linker when built via the Makefile.
package version

var (
	Version   = "0.1.3"
	BuildTime = "unknown"
	GitCommit = "unknown"
)
