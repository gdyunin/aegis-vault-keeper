package note

import (
	"testing"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoteErrRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		errorIn        error
		name           string
		expectedMsg    string
		expectedStatus int
		expectedClass  errutil.ErrorClass
		expectedLogIt  bool
		expectedMerge  bool
	}{
		{
			name:           "tech error",
			errorIn:        app.ErrNoteTechError,
			expectedStatus: 500,
			expectedMsg:    "Internal Server Error",
			expectedLogIt:  true,
			expectedMerge:  false,
			expectedClass:  errutil.ErrorClassTech,
		},
		{
			name:           "access denied error",
			errorIn:        app.ErrNoteAccessDenied,
			expectedStatus: 403,
			expectedMsg:    "Access to this note is denied",
			expectedLogIt:  false,
			expectedMerge:  false,
			expectedClass:  errutil.ErrorClassAuth,
		},
		{
			name:           "not found error",
			errorIn:        app.ErrNoteNotFound,
			expectedStatus: 404,
			expectedMsg:    "Note not found",
			expectedLogIt:  false,
			expectedMerge:  false,
			expectedClass:  errutil.ErrorClassGeneric,
		},
		{
			name:           "incorrect note text error",
			errorIn:        app.ErrNoteIncorrectNoteText,
			expectedStatus: 400,
			expectedMsg:    "Invalid note text",
			expectedLogIt:  false,
			expectedMerge:  true,
			expectedClass:  errutil.ErrorClassValidation,
		},
		{
			name:           "app error",
			errorIn:        app.ErrNoteAppError,
			expectedStatus: 400,
			expectedMsg:    "Invalid parameters",
			expectedLogIt:  false,
			expectedMerge:  false,
			expectedClass:  errutil.ErrorClassValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Find the error in the registry
			var found bool
			var policy errutil.Policy
			for _, rule := range NoteErrRegistry {
				if rule.ErrorIn == tt.errorIn {
					found = true
					policy = rule.HandlePolicy
					break
				}
			}

			require.True(t, found, "Error should be found in registry")
			assert.Equal(t, tt.expectedStatus, policy.StatusCode)
			assert.Equal(t, tt.expectedMsg, policy.PublicMsg)
			assert.Equal(t, tt.expectedLogIt, policy.LogIt)
			assert.Equal(t, tt.expectedMerge, policy.AllowMerge)
			assert.Equal(t, tt.expectedClass, policy.ErrorClass)
		})
	}
}

func TestNoteErrRegistry_Coverage(t *testing.T) {
	t.Parallel()

	// Verify all expected note errors are covered
	expectedErrors := []error{
		app.ErrNoteTechError,
		app.ErrNoteAccessDenied,
		app.ErrNoteNotFound,
		app.ErrNoteIncorrectNoteText,
		app.ErrNoteAppError,
	}

	registryErrors := make(map[error]bool)
	for _, rule := range NoteErrRegistry {
		registryErrors[rule.ErrorIn] = true
	}

	for _, expectedErr := range expectedErrors {
		assert.True(t, registryErrors[expectedErr], "Error %v should be in registry", expectedErr)
	}

	// Verify registry has exactly the expected number of rules
	assert.Len(t, NoteErrRegistry, len(expectedErrors))
}

func TestNoteErrRegistry_PolicyConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		errorClass  errutil.ErrorClass
		expectLogIt bool
	}{
		{
			name:        "tech errors should be logged",
			errorClass:  errutil.ErrorClassTech,
			expectLogIt: true,
		},
		{
			name:        "auth errors should not be logged",
			errorClass:  errutil.ErrorClassAuth,
			expectLogIt: false,
		},
		{
			name:        "validation errors should not be logged",
			errorClass:  errutil.ErrorClassValidation,
			expectLogIt: false,
		},
		{
			name:        "generic errors should not be logged",
			errorClass:  errutil.ErrorClassGeneric,
			expectLogIt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			for _, rule := range NoteErrRegistry {
				if rule.HandlePolicy.ErrorClass == tt.errorClass {
					assert.Equal(t, tt.expectLogIt, rule.HandlePolicy.LogIt,
						"Error class %s should have LogIt=%v", tt.errorClass, tt.expectLogIt)
				}
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	tests := []struct {
		inputError     error
		name           string
		expectedMsgs   []string
		expectedStatus int
	}{
		{
			name:           "note tech error",
			inputError:     app.ErrNoteTechError,
			expectedStatus: 500,
			expectedMsgs:   []string{"Internal Server Error"},
		},
		{
			name:           "note access denied",
			inputError:     app.ErrNoteAccessDenied,
			expectedStatus: 403,
			expectedMsgs:   []string{"Access to this note is denied"},
		},
		{
			name:           "note not found",
			inputError:     app.ErrNoteNotFound,
			expectedStatus: 404,
			expectedMsgs:   []string{"Note not found"},
		},
		{
			name:           "incorrect note text",
			inputError:     app.ErrNoteIncorrectNoteText,
			expectedStatus: 400,
			expectedMsgs:   []string{"Invalid note text"},
		},
		{
			name:           "note app error",
			inputError:     app.ErrNoteAppError,
			expectedStatus: 400,
			expectedMsgs:   []string{"Invalid parameters"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c, _ := gin.CreateTestContext(nil)

			status, msgs := handleError(tt.inputError, c)

			assert.Equal(t, tt.expectedStatus, status)
			assert.Equal(t, tt.expectedMsgs, msgs)
		})
	}
}

func TestNoteErrRegistry_StatusCodes(t *testing.T) {
	t.Parallel()

	// Verify that status codes are within expected ranges
	for _, rule := range NoteErrRegistry {
		assert.True(t, rule.HandlePolicy.StatusCode >= 400 && rule.HandlePolicy.StatusCode < 600,
			"Status code should be a valid HTTP error code (4xx or 5xx)")
	}
}

func TestNoteErrRegistry_Messages(t *testing.T) {
	t.Parallel()

	// Verify that all messages are non-empty
	for _, rule := range NoteErrRegistry {
		assert.NotEmpty(t, rule.HandlePolicy.PublicMsg,
			"Public message should not be empty")
	}
}
