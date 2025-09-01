package buildinfo

// Version holds the application version string.
// Set at build time via -ldflags "-X ...".
var Version = "N/A"

// Date holds the build timestamp in RFC3339 format.
// Set at build time via -ldflags "-X ...".
var Date = "N/A"

// Commit holds the Git commit hash used for the build.
// Set at build time via -ldflags "-X ...".
var Commit = "N/A"
