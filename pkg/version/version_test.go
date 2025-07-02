package version

import (
	"testing"
)

func TestMajorVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    int
	}{
		{"standard semver", "1.2.3", 1},
		{"double digit major", "12.0.0", 12},
		{"zero major", "0.1.0", 0},
		{"with pre-release", "2.0.0-alpha", 2},
		{"with build metadata", "3.0.0+build123", 3},
		{"invalid format", "invalid", 0},
		{"empty string", "", 0},
		{"just major", "5", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original and restore after test
			original := Version
			defer func() { Version = original }()

			Version = tt.version
			if got := MajorVersion(); got != tt.want {
				t.Errorf("MajorVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinorVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    int
	}{
		{"standard semver", "1.2.3", 2},
		{"double digit minor", "1.12.0", 12},
		{"zero minor", "1.0.3", 0},
		{"with pre-release", "1.3.0-beta", 3},
		{"missing minor", "1", 0},
		{"invalid format", "invalid", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := Version
			defer func() { Version = original }()

			Version = tt.version
			if got := MinorVersion(); got != tt.want {
				t.Errorf("MinorVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatchVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    int
	}{
		{"standard semver", "1.2.3", 3},
		{"double digit patch", "1.2.12", 12},
		{"zero patch", "1.2.0", 0},
		{"with pre-release", "1.2.3-rc1", 3},
		{"with pre-release dash", "1.2.3-alpha-1", 3},
		{"missing patch", "1.2", 0},
		{"only major", "1", 0},
		{"invalid format", "invalid", 0},
		{"empty string", "", 0},
		{"git describe format", "1.2.3-7-gabc1234", 3},
		{"git describe dirty", "1.2.3-7-gabc1234-dirty", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := Version
			defer func() { Version = original }()

			Version = tt.version
			if got := PatchVersion(); got != tt.want {
				t.Errorf("PatchVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	// Save originals and restore after test
	originalVersion := Version
	originalCommit := GitCommit
	originalDate := BuildDate
	defer func() {
		Version = originalVersion
		GitCommit = originalCommit
		BuildDate = originalDate
	}()

	// Set test values
	Version = "1.2.3"
	GitCommit = "abc1234"
	BuildDate = "2024-01-01T00:00:00Z"

	info := Info()

	if info.Version != "1.2.3" {
		t.Errorf("Info().Version = %v, want %v", info.Version, "1.2.3")
	}
	if info.Major != 1 {
		t.Errorf("Info().Major = %v, want %v", info.Major, 1)
	}
	if info.Minor != 2 {
		t.Errorf("Info().Minor = %v, want %v", info.Minor, 2)
	}
	if info.Patch != 3 {
		t.Errorf("Info().Patch = %v, want %v", info.Patch, 3)
	}
	if info.GitCommit != "abc1234" {
		t.Errorf("Info().GitCommit = %v, want %v", info.GitCommit, "abc1234")
	}
	if info.BuildDate != "2024-01-01T00:00:00Z" {
		t.Errorf("Info().BuildDate = %v, want %v", info.BuildDate, "2024-01-01T00:00:00Z")
	}
}

func TestVersionWithGitDescribe(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		wantMajor   int
		wantMinor   int
		wantPatch   int
		description string
	}{
		{
			name:        "clean tag",
			version:     "1.2.3",
			wantMajor:   1,
			wantMinor:   2,
			wantPatch:   3,
			description: "Git tag exactly on release",
		},
		{
			name:        "commits after tag",
			version:     "1.2.3-7-gabc1234",
			wantMajor:   1,
			wantMinor:   2,
			wantPatch:   3,
			description: "7 commits after v1.2.3",
		},
		{
			name:        "dirty working directory",
			version:     "1.2.3-7-gabc1234-dirty",
			wantMajor:   1,
			wantMinor:   2,
			wantPatch:   3,
			description: "Uncommitted changes",
		},
		{
			name:        "pre-release version",
			version:     "2.0.0-beta.1",
			wantMajor:   2,
			wantMinor:   0,
			wantPatch:   0,
			description: "Beta release",
		},
		{
			name:        "no tags yet",
			version:     "0.0.0-1-gabc1234",
			wantMajor:   0,
			wantMinor:   0,
			wantPatch:   0,
			description: "Repository with no tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := Version
			defer func() { Version = original }()

			Version = tt.version
			
			if got := MajorVersion(); got != tt.wantMajor {
				t.Errorf("MajorVersion() = %v, want %v for %s", got, tt.wantMajor, tt.description)
			}
			if got := MinorVersion(); got != tt.wantMinor {
				t.Errorf("MinorVersion() = %v, want %v for %s", got, tt.wantMinor, tt.description)
			}
			if got := PatchVersion(); got != tt.wantPatch {
				t.Errorf("PatchVersion() = %v, want %v for %s", got, tt.wantPatch, tt.description)
			}
		})
	}
}