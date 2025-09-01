package common

import (
	"testing"
	"time"
)

// TestNewBuildInfoOperator tests the NewBuildInfoOperator constructor function.
func TestNewBuildInfoOperator(t *testing.T) {
	tests := []struct {
		name    string
		version string
		date    string
		commit  string
	}{
		{
			name:    "valid build info",
			version: "v1.0.0",
			date:    "2023-01-01T12:00:00Z",
			commit:  "abc123",
		},
		{
			name:    "empty values",
			version: "",
			date:    "",
			commit:  "",
		},
		{
			name:    "semantic version with prerelease",
			version: "v2.1.0-beta.1",
			date:    "2023-08-09T10:30:00Z",
			commit:  "1234567890abcdef",
		},
		{
			name:    "full SHA-1 commit hash",
			version: "v3.0.0",
			date:    "2023-12-31T23:59:59Z",
			commit:  "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operator := NewBuildInfoOperator(tt.version, tt.date, tt.commit)

			// Ensure operator is not nil before accessing its fields
			if operator == nil {
				t.Fatal("NewBuildInfoOperator returned nil")
				return // This return is unreachable after Fatal, but helps static analysis
			}

			if operator.buildVersion != tt.version {
				t.Errorf("Expected buildVersion %s, got %s", tt.version, operator.buildVersion)
			}

			if operator.buildDate != tt.date {
				t.Errorf("Expected buildDate %s, got %s", tt.date, operator.buildDate)
			}

			if operator.buildCommit != tt.commit {
				t.Errorf("Expected buildCommit %s, got %s", tt.commit, operator.buildCommit)
			}
		})
	}
}

// TestBuildInfoOperator_Version tests the Version method.
func TestBuildInfoOperator_Version(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{
			name:     "valid version",
			version:  "v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:     "empty version",
			version:  "",
			expected: "",
		},
		{
			name:     "semantic version",
			version:  "1.2.3-alpha.1",
			expected: "1.2.3-alpha.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operator := NewBuildInfoOperator(tt.version, "2023-01-01T12:00:00Z", "abc123")
			result := operator.Version()

			if result != tt.expected {
				t.Errorf("Version() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestBuildInfoOperator_Date tests the Date method parsing and time conversion.
func TestBuildInfoOperator_Date(t *testing.T) {
	tests := []struct {
		expectedTime time.Time
		name         string
		buildDate    string
		shouldBeZero bool
	}{
		{
			name:         "valid RFC3339 date",
			buildDate:    "2023-01-01T12:00:00Z",
			expectedTime: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			shouldBeZero: false,
		},
		{
			name:         "valid RFC3339 date with timezone",
			buildDate:    "2023-12-25T15:30:45+03:00",
			expectedTime: time.Date(2023, 12, 25, 12, 30, 45, 0, time.UTC),
			shouldBeZero: false,
		},
		{
			name:         "invalid date format",
			buildDate:    "2023-01-01",
			expectedTime: time.Time{},
			shouldBeZero: true,
		},
		{
			name:         "empty date",
			buildDate:    "",
			expectedTime: time.Time{},
			shouldBeZero: true,
		},
		{
			name:         "malformed date",
			buildDate:    "not-a-date",
			expectedTime: time.Time{},
			shouldBeZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operator := NewBuildInfoOperator("v1.0.0", tt.buildDate, "abc123")
			result := operator.Date()

			if tt.shouldBeZero {
				if !result.IsZero() {
					t.Errorf("Date() = %v, expected zero time", result)
				}
			} else {
				if !result.Equal(tt.expectedTime) {
					t.Errorf("Date() = %v, want %v", result, tt.expectedTime)
				}
			}
		})
	}
}

// TestBuildInfoOperator_Commit tests the Commit method.
func TestBuildInfoOperator_Commit(t *testing.T) {
	tests := []struct {
		name     string
		commit   string
		expected string
	}{
		{
			name:     "valid commit hash",
			commit:   "abc123def456",
			expected: "abc123def456",
		},
		{
			name:     "short commit hash",
			commit:   "abc123",
			expected: "abc123",
		},
		{
			name:     "empty commit",
			commit:   "",
			expected: "",
		},
		{
			name:     "full SHA-1 hash",
			commit:   "da39a3ee5e6b4b0d3255bfef95601890afd80709",
			expected: "da39a3ee5e6b4b0d3255bfef95601890afd80709",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operator := NewBuildInfoOperator("v1.0.0", "2023-01-01T12:00:00Z", tt.commit)
			result := operator.Commit()

			if result != tt.expected {
				t.Errorf("Commit() = %v, want %v", result, tt.expected)
			}
		})
	}
}
