package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildInfoVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		variable *string
		validate func(t *testing.T, value string)
		name     string
	}{
		{
			name:     "Version variable exists",
			variable: &Version,
			validate: func(t *testing.T, value string) {
				t.Helper()
				assert.NotEmpty(t, value)
			},
		},
		{
			name:     "Date variable exists",
			variable: &Date,
			validate: func(t *testing.T, value string) {
				t.Helper()
				assert.NotEmpty(t, value)
			},
		},
		{
			name:     "Commit variable exists",
			variable: &Commit,
			validate: func(t *testing.T, value string) {
				t.Helper()
				assert.NotEmpty(t, value)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.validate(t, *tt.variable)
		})
	}
}

func TestBuildInfoVariableTypes(t *testing.T) {
	t.Parallel()

	// Test that all variables are strings
	assert.IsType(t, "", Version, "Version should be a string")
	assert.IsType(t, "", Date, "Date should be a string")
	assert.IsType(t, "", Commit, "Commit should be a string")
}

func TestBuildInfoVariableModification(t *testing.T) {
	t.Parallel()

	// Store original values
	originalVersion := Version
	originalDate := Date
	originalCommit := Commit

	// Modify variables
	Version = "v1.0.0"
	Date = "2023-01-01T12:00:00Z"
	Commit = "abc123"

	// Test modified values
	assert.Equal(t, "v1.0.0", Version)
	assert.Equal(t, "2023-01-01T12:00:00Z", Date)
	assert.Equal(t, "abc123", Commit)

	// Restore original values
	Version = originalVersion
	Date = originalDate
	Commit = originalCommit

	// Verify restoration
	assert.Equal(t, originalVersion, Version)
	assert.Equal(t, originalDate, Date)
	assert.Equal(t, originalCommit, Commit)
}
