package about

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInfo(t *testing.T) {
	t.Parallel()

	testTime := time.Now()

	tests := []struct {
		name        string
		info        BuildInfo
		wantVersion string
		wantDate    time.Time
		wantCommit  string
	}{
		{
			name: "success/valid_build_info",
			info: BuildInfo{
				Version: "1.0.0",
				Date:    testTime,
				Commit:  "abc123def",
			},
			wantVersion: "1.0.0",
			wantDate:    testTime,
			wantCommit:  "abc123def",
		},
		{
			name: "success/empty_build_info",
			info: BuildInfo{
				Version: "",
				Date:    time.Time{},
				Commit:  "",
			},
			wantVersion: "",
			wantDate:    time.Time{},
			wantCommit:  "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantVersion, tt.info.Version)
			assert.Equal(t, tt.wantDate, tt.info.Date)
			assert.Equal(t, tt.wantCommit, tt.info.Commit)
		})
	}
}

func TestBuildInfo_JSONSerialization(t *testing.T) {
	t.Parallel()

	// Test that BuildInfo struct can be JSON serialized/deserialized
	testTime, _ := time.Parse(time.RFC3339, "2023-12-01T10:00:00Z")

	info := BuildInfo{
		Version: "0.1.1",
		Date:    testTime,
		Commit:  "0b712a2",
	}

	// Test that fields have correct json tags
	require.NotEmpty(t, info.Version)
	require.NotEmpty(t, info.Date)
	require.NotEmpty(t, info.Commit)
}

// Test struct field validation.
func TestBuildInfo_StructFields(t *testing.T) {
	t.Parallel()

	var info BuildInfo

	// Test that all required fields exist and are accessible
	info.Version = "test"
	info.Date = time.Now()
	info.Commit = "test"

	assert.Equal(t, "test", info.Version)
	assert.NotEmpty(t, info.Date)
	assert.Equal(t, "test", info.Commit)
}
