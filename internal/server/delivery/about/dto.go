package about

import "time"

// BuildInfo represents application build information.
type BuildInfo struct {
	// Version is the semantic version string of the application build.
	Version string `json:"version" example:"0.1.1"` // Application version
	// Date is the timestamp when the application was built.
	Date time.Time `json:"date"    example:"2023-12-01T10:00:00Z"` // Build date
	// Commit is the Git commit hash from which the application was built.
	Commit string `json:"commit"  example:"0b712a2"` // Git commit hash
}
