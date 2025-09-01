package common

import "time"

// BuildInfoOperator provides access to application build information and metadata.
type BuildInfoOperator struct {
	// buildVersion contains the application version string.
	buildVersion string
	// buildDate contains the build timestamp in RFC3339 format.
	buildDate string
	// buildCommit contains the Git commit hash used for the build.
	buildCommit string
}

// NewBuildInfoOperator creates a new BuildInfoOperator with the provided build metadata.
func NewBuildInfoOperator(version, date, commit string) *BuildInfoOperator {
	return &BuildInfoOperator{
		buildVersion: version,
		buildDate:    date,
		buildCommit:  commit,
	}
}

// Version returns the application version string.
func (b *BuildInfoOperator) Version() string {
	return b.buildVersion
}

// Date returns the build timestamp as a parsed time.Time value.
// Returns zero time if the build date cannot be parsed.
func (b *BuildInfoOperator) Date() time.Time {
	parsedDate, err := time.Parse(time.RFC3339, b.buildDate)
	if err != nil {
		return time.Time{}
	}
	return parsedDate
}

// Commit returns the Git commit hash used for the build.
func (b *BuildInfoOperator) Commit() string {
	return b.buildCommit
}
