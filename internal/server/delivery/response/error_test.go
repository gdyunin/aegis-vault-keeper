package response

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestError_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		error    Error
		expected []string
	}{
		{
			name: "single message",
			error: Error{
				Messages: []string{"Single error message"},
			},
			expected: []string{"Single error message"},
		},
		{
			name: "multiple messages",
			error: Error{
				Messages: []string{"First error", "Second error", "Third error"},
			},
			expected: []string{"First error", "Second error", "Third error"},
		},
		{
			name: "empty messages",
			error: Error{
				Messages: []string{},
			},
			expected: []string{},
		},
		{
			name: "nil messages",
			error: Error{
				Messages: nil,
			},
			expected: nil,
		},
		{
			name: "empty string message",
			error: Error{
				Messages: []string{""},
			},
			expected: []string{""},
		},
		{
			name: "mixed empty and non-empty messages",
			error: Error{
				Messages: []string{"Valid message", "", "Another valid message"},
			},
			expected: []string{"Valid message", "", "Another valid message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.error.Messages)
		})
	}
}

func TestError_JSONSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		expectedJSON string
		error        Error
	}{
		{
			name: "single message",
			error: Error{
				Messages: []string{"Test error"},
			},
			expectedJSON: `{"messages":["Test error"]}`,
		},
		{
			name: "multiple messages",
			error: Error{
				Messages: []string{"Error 1", "Error 2"},
			},
			expectedJSON: `{"messages":["Error 1","Error 2"]}`,
		},
		{
			name: "empty messages",
			error: Error{
				Messages: []string{},
			},
			expectedJSON: `{"messages":[]}`,
		},
		{
			name: "nil messages",
			error: Error{
				Messages: nil,
			},
			expectedJSON: `{"messages":null}`,
		},
		{
			name: "special characters",
			error: Error{
				Messages: []string{"Error with \"quotes\" and \\backslashes"},
			},
			expectedJSON: `{"messages":["Error with \"quotes\" and \\backslashes"]}`,
		},
		{
			name: "unicode characters",
			error: Error{
				Messages: []string{"错误信息", "エラーメッセージ", "🚫 Error"},
			},
			expectedJSON: `{"messages":["错误信息","エラーメッセージ","🚫 Error"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			jsonData, err := json.Marshal(tt.error)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expectedJSON, string(jsonData))
		})
	}
}

func TestError_JSONDeserialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		jsonData    string
		expected    Error
		expectError bool
	}{
		{
			name:     "single message",
			jsonData: `{"messages":["Test error"]}`,
			expected: Error{
				Messages: []string{"Test error"},
			},
		},
		{
			name:     "multiple messages",
			jsonData: `{"messages":["Error 1","Error 2"]}`,
			expected: Error{
				Messages: []string{"Error 1", "Error 2"},
			},
		},
		{
			name:     "empty messages",
			jsonData: `{"messages":[]}`,
			expected: Error{
				Messages: []string{},
			},
		},
		{
			name:     "null messages",
			jsonData: `{"messages":null}`,
			expected: Error{
				Messages: nil,
			},
		},
		{
			name:        "invalid json",
			jsonData:    `{"messages":["unclosed`,
			expectError: true,
		},
		{
			name:        "wrong type for messages",
			jsonData:    `{"messages":"not an array"}`,
			expectError: true,
		},
		{
			name:     "extra fields ignored",
			jsonData: `{"messages":["Test"],"extra":"ignored"}`,
			expected: Error{
				Messages: []string{"Test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var result Error
			err := json.Unmarshal([]byte(tt.jsonData), &result)

			if tt.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultBadRequestError(t *testing.T) {
	// Note: Not using t.Parallel() because we're testing global variables

	// Test the default bad request error
	assert.NotNil(t, DefaultBadRequestError)
	assert.Len(t, DefaultBadRequestError.Messages, 1)
	expectedMessage := http.StatusText(http.StatusBadRequest)
	assert.Equal(t, expectedMessage, DefaultBadRequestError.Messages[0])
	assert.Equal(t, "Bad Request", DefaultBadRequestError.Messages[0])

	// Test JSON serialization
	jsonData, err := json.Marshal(DefaultBadRequestError)
	require.NoError(t, err)
	assert.JSONEq(t, `{"messages":["Bad Request"]}`, string(jsonData))
}

func TestDefaultInternalServerError(t *testing.T) {
	// Note: Not using t.Parallel() because we're testing global variables

	// Test the default internal server error
	assert.NotNil(t, DefaultInternalServerError)
	assert.Len(t, DefaultInternalServerError.Messages, 1)
	expectedMessage := http.StatusText(http.StatusInternalServerError)
	assert.Equal(t, expectedMessage, DefaultInternalServerError.Messages[0])
	assert.Equal(t, "Internal Server Error", DefaultInternalServerError.Messages[0])

	// Test JSON serialization
	jsonData, err := json.Marshal(DefaultInternalServerError)
	require.NoError(t, err)
	assert.JSONEq(t, `{"messages":["Internal Server Error"]}`, string(jsonData))
}

func TestDefaultErrors_Immutability(t *testing.T) {
	// Note: Not using t.Parallel() because we're testing global variables

	// Test that the default errors have the expected values
	// These are global variables, so we can't actually test immutability
	// We can only test that they have the expected initial values

	assert.NotNil(t, DefaultBadRequestError)
	assert.Len(t, DefaultBadRequestError.Messages, 1)
	assert.Equal(t, "Bad Request", DefaultBadRequestError.Messages[0])

	assert.NotNil(t, DefaultInternalServerError)
	assert.Len(t, DefaultInternalServerError.Messages, 1)
	assert.Equal(t, "Internal Server Error", DefaultInternalServerError.Messages[0])

	// Test that we can create independent copies
	badRequestCopy := Error{
		Messages: []string{DefaultBadRequestError.Messages[0]},
	}
	badRequestCopy.Messages[0] = "Modified Bad Request"

	internalServerCopy := Error{
		Messages: []string{DefaultInternalServerError.Messages[0]},
	}
	internalServerCopy.Messages[0] = "Modified Internal Server Error"

	// Verify copies are modified
	assert.Equal(t, "Modified Bad Request", badRequestCopy.Messages[0])
	assert.Equal(t, "Modified Internal Server Error", internalServerCopy.Messages[0])
}

func TestError_RoundTripSerialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		error Error
	}{
		{
			name: "single message round trip",
			error: Error{
				Messages: []string{"Round trip test"},
			},
		},
		{
			name: "multiple messages round trip",
			error: Error{
				Messages: []string{"Message 1", "Message 2", "Message 3"},
			},
		},
		{
			name: "empty messages round trip",
			error: Error{
				Messages: []string{},
			},
		},
		{
			name: "special characters round trip",
			error: Error{
				Messages: []string{"Special: !@#$%^&*()", "Unicode: 测试 🚀"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Serialize to JSON
			jsonData, err := json.Marshal(tt.error)
			require.NoError(t, err)

			// Deserialize back to struct
			var result Error
			err = json.Unmarshal(jsonData, &result)
			require.NoError(t, err)

			// Verify round trip
			assert.Equal(t, tt.error, result)
		})
	}
}

func TestError_EdgeCases(t *testing.T) {
	t.Parallel()

	// Test with very long message
	longMessage := make([]byte, 10000)
	for i := range longMessage {
		longMessage[i] = 'a'
	}

	longError := Error{
		Messages: []string{string(longMessage)},
	}

	jsonData, err := json.Marshal(longError)
	require.NoError(t, err)

	var result Error
	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)
	assert.Equal(t, longError, result)

	// Test with many messages
	manyMessages := make([]string, 1000)
	for i := range manyMessages {
		manyMessages[i] = "Message " + string(rune('0'+i%10))
	}

	manyError := Error{
		Messages: manyMessages,
	}

	jsonData, err = json.Marshal(manyError)
	require.NoError(t, err)

	err = json.Unmarshal(jsonData, &result)
	require.NoError(t, err)
	assert.Equal(t, manyError, result)
}
