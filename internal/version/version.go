// Package version provides access to version information set at build time.
package version

import (
	"os/exec"
	"strings"
	"time"
)

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

// GetRuntimeVersionInfo 获取运行时的版本信息
func GetRuntimeVersionInfo() (version, buildTime, commitSHA string) {
	// 如果通过 -ldflags 注入了值，直接返回
	if Version != "dev" || BuildTime != "unknown" || CommitSHA != "unknown" {
		return Version, BuildTime, CommitSHA
	}

	// 默认值
	version = "dev"
	commitSHA = "unknown"

	// 尝试从 git 获取信息
	if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		version = strings.TrimSpace(string(out))
	}

	if out, err := exec.Command("git", "rev-parse", "--short", "HEAD").Output(); err == nil {
		commitSHA = strings.TrimSpace(string(out))
	}

	// 使用当前时间
	buildTime = time.Now().Format("2006-01-02 15:04:05")

	return
}
