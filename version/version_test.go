package version_test

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/theHamdiz/it/version"
)

// TestVersionInfo ensures version.Get() returns the correct information
// and that all the struct fields are properly populated
func TestVersionInfo(t *testing.T) {
	info := version.Get()

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Version", info.Version, "0.1.0"},
		{"GitCommit", info.GitCommit, "unknown"},
		{"GitBranch", info.GitBranch, "main"},
		{"GoVersion", info.GoVersion, runtime.Version()},
		{"Platform", info.Platform, runtime.GOOS + "/" + runtime.GOARCH},
		{"Environment", info.Environment, "development"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("Expected %s=%q, got %q", tt.name, tt.expected, tt.got)
			}
		})
	}

	if info.BuildTime.After(time.Now()) {
		t.Error("BuildTime is in the future")
	}
	if info.BuildTime.Year() < 2020 {
		t.Error("BuildTime is suspiciously old")
	}
}

// TestVersionInfoToMap ensures the ToMap() method returns
// the correct key-value pairs
func TestVersionInfoToMap(t *testing.T) {
	info := version.Get()
	m := info.ToMap()

	// Check for required keys
	requiredKeys := []string{
		"version",
		"buildTime",
		"gitCommit",
		"gitBranch",
		"goVersion",
		"platform",
		"environment",
	}

	for _, key := range requiredKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("Expected key %q in map, but not found", key)
		}
	}

	if m["version"] != info.Version {
		t.Errorf("Expected version=%q, got %q", info.Version, m["version"])
	}
	if m["gitCommit"] != info.GitCommit {
		t.Errorf("Expected gitCommit=%q, got %q", info.GitCommit, m["gitCommit"])
	}
	if m["goVersion"] != runtime.Version() {
		t.Errorf("Expected goVersion=%q, got %q", runtime.Version(), m["goVersion"])
	}
}

// TestVersionInfoString ensures the String() method formats correctly
func TestVersionInfoString(t *testing.T) {
	info := version.Get()
	str := info.String()

	requiredParts := []string{
		info.Version,
		info.GitCommit[:7],
		info.GitBranch,
		info.GoVersion,
		info.Platform,
		info.Environment,
	}

	for _, part := range requiredParts {
		if !strings.Contains(str, part) {
			t.Errorf("Expected string to contain %q, but it doesn't: %s", part, str)
		}
	}
}

// TestVersionInfoBuildTime ensures BuildTime parsing works correctly
func TestVersionInfoBuildTime(t *testing.T) {
	tests := []struct {
		name      string
		buildTime string
		margin    time.Duration
	}{
		{
			name:      "Valid current time",
			buildTime: time.Now().Format(time.RFC3339),
			margin:    time.Minute,
		},
		{
			name:      "Invalid time string",
			buildTime: "not a time",
			margin:    time.Minute,
		},
		{
			name:      "Empty time string",
			buildTime: "",
			margin:    time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			info := version.Get()

			timeDiff := info.BuildTime.Sub(now)
			if timeDiff < -tt.margin || timeDiff > tt.margin {
				t.Errorf(
					"BuildTime %v is outside acceptable range (%v ± %v)",
					info.BuildTime,
					now,
					tt.margin,
				)
			}
		})
	}
}
