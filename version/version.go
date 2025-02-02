// Package version - Because semantic versioning is cheaper than therapy
package version

import (
	"fmt"
	"runtime"
	"time"
)

// Info holds all the version information in one neat package
// Like a CVS but for your software's identity crisis
type Info struct {
	Version     string    // Semantic version (or whatever we feel like)
	BuildTime   time.Time // When this all started
	GitCommit   string    // The commit we're blaming
	GitBranch   string    // Which branch caused this mess
	GoVersion   string    // Which Go version to complain about
	Platform    string    // What we're running on
	Environment string    // prod/staging/dev/whoknows
}

var (
	// These get set during build time via ldflags
	version     = "0.1.0"
	buildTime   = ""
	gitCommit   = "unknown"
	gitBranch   = "main"
	environment = "development"
)

// Get returns structured version info that's actually useful
func Get() Info {
	// Parse build time or default to now
	bt, err := time.Parse(time.RFC3339, buildTime)
	if err != nil {
		bt = time.Now()
	}

	return Info{
		Version:     version,
		BuildTime:   bt,
		GitCommit:   gitCommit,
		GitBranch:   gitBranch,
		GoVersion:   runtime.Version(),
		Platform:    fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Environment: environment,
	}
}

// String returns a human-readable version string
// For when you need to print it all in one line
func (i Info) String() string {
	return fmt.Sprintf(
		"%s built at %s from %s@%s (%s) running on %s in %s",
		i.Version,
		i.BuildTime.Format(time.RFC3339),
		i.GitBranch,
		i.GitCommit[:7], // Short SHA
		i.GoVersion,
		i.Platform,
		i.Environment,
	)
}

// ToMap converts Info to a map[string]string
// For APIs that prefer key-value pairs
func (i Info) ToMap() map[string]string {
	return map[string]string{
		"version":     i.Version,
		"buildTime":   i.BuildTime.Format(time.RFC3339),
		"gitCommit":   i.GitCommit,
		"gitBranch":   i.GitBranch,
		"goVersion":   i.GoVersion,
		"platform":    i.Platform,
		"environment": i.Environment,
	}
}

// Build command would look like:
// go build -ldflags="-X 'package/version.version=1.0.0' \
//                    -X 'package/version.buildTime=$(date -u +"%Y-%m-%dT%H:%M:%SZ")' \
//                    -X 'package/version.gitCommit=$(git rev-parse HEAD)' \
//                    -X 'package/version.gitBranch=$(git rev-parse --abbrev-ref HEAD)' \
//                    -X 'package/version.environment=production'"

// Usage example:
/*
func main() {
    info := version.Get()

    // Print full info
    fmt.Println(info)

    // Access specific fields
    fmt.Printf("Version: %s\n", info.Version)
    fmt.Printf("Built: %s\n", info.BuildTime.Format("2006-01-02"))

    // Use as map
    infoMap := info.ToMap()
    json.NewEncoder(os.Stdout).Encode(infoMap)
}
*/
