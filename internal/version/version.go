// Package version holds the application version.
//
// Release builds inject it with
//
//	go build -ldflags "-X lantransfer/internal/version.Version=2.0.0"
package version

// Version is "dev" for local builds without -ldflags.
var Version = "dev"

// String returns the version prefixed with "v" for display ("v2.0.0", "dev").
func String() string {
	if Version == "" || Version == "dev" {
		return "dev"
	}
	return "v" + Version
}
