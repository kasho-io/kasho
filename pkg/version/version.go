package version

import (
	"strconv"
	"strings"
)

var (
	// Version is the full semver version (set at build time)
	Version = "0.0.0"

	// GitCommit is the git commit hash (set at build time)
	GitCommit = "unknown"

	// BuildDate is the build timestamp (set at build time)
	BuildDate = "unknown"
)

// MajorVersion returns the major version number
func MajorVersion() int {
	parts := strings.Split(Version, ".")
	if len(parts) > 0 {
		major, _ := strconv.Atoi(parts[0])
		return major
	}
	return 0
}

// MinorVersion returns the minor version number
func MinorVersion() int {
	parts := strings.Split(Version, ".")
	if len(parts) > 1 {
		minor, _ := strconv.Atoi(parts[1])
		return minor
	}
	return 0
}

// PatchVersion returns the patch version number
func PatchVersion() int {
	parts := strings.Split(Version, ".")
	if len(parts) > 2 {
		// Handle versions with pre-release info (e.g., "1.2.3-alpha")
		patchPart := strings.Split(parts[2], "-")[0]
		patch, _ := strconv.Atoi(patchPart)
		return patch
	}
	return 0
}

// Info returns structured version information
func Info() VersionInfo {
	return VersionInfo{
		Version:   Version,
		Major:     MajorVersion(),
		Minor:     MinorVersion(),
		Patch:     PatchVersion(),
		GitCommit: GitCommit,
		BuildDate: BuildDate,
	}
}

// VersionInfo contains structured version information
type VersionInfo struct {
	Version   string
	Major     int
	Minor     int
	Patch     int
	GitCommit string
	BuildDate string
}