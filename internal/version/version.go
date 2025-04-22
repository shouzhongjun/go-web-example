// Package version provides access to version information set at build time.
package version

// Version information set by build flags
var (
	// Version is the current version of the application
	Version = "dev"
	// BuildTime is the time the application was built
	BuildTime = "unknown"
	// CommitSHA is the git commit SHA at the time of the build
	CommitSHA = "unknown"
)

// GetVersion returns the current version information
func GetVersion() string {
	return Version
}

// GetBuildTime returns the build time
func GetBuildTime() string {
	return BuildTime
}

// GetCommitSHA returns the git commit SHA
func GetCommitSHA() string {
	return CommitSHA
}

// GetVersionInfo returns all version information as a map
func GetVersionInfo() map[string]string {
	return map[string]string{
		"version":   Version,
		"buildTime": BuildTime,
		"commitSHA": CommitSHA,
	}
}
