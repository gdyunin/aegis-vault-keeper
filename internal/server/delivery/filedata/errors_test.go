package filedata

import (
	"testing"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileDataErrRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		expectedError error
		name          string
		expectedClass errutil.ErrorClass
		expectedCode  int
	}{
		{
			name:          "ErrFileTechError",
			expectedError: app.ErrFileTechError,
			expectedClass: errutil.ErrorClassTech,
			expectedCode:  500,
		},
		{
			name:          "ErrFileAccessDenied",
			expectedError: app.ErrFileAccessDenied,
			expectedClass: errutil.ErrorClassAuth,
			expectedCode:  403,
		},
		{
			name:          "ErrFileNotFound",
			expectedError: app.ErrFileNotFound,
			expectedClass: errutil.ErrorClassGeneric,
			expectedCode:  404,
		},
		{
			name:          "ErrFileIncorrectStorageKey",
			expectedError: app.ErrFileIncorrectStorageKey,
			expectedClass: errutil.ErrorClassValidation,
			expectedCode:  400,
		},
		{
			name:          "ErrFileIncorrectHashSum",
			expectedError: app.ErrFileIncorrectHashSum,
			expectedClass: errutil.ErrorClassValidation,
			expectedCode:  400,
		},
		{
			name:          "ErrFileDataRequired",
			expectedError: app.ErrFileDataRequired,
			expectedClass: errutil.ErrorClassValidation,
			expectedCode:  400,
		},
		{
			name:          "ErrRollBackFileSaveFailed",
			expectedError: app.ErrRollBackFileSaveFailed,
			expectedClass: errutil.ErrorClassTech,
			expectedCode:  500,
		},
		{
			name:          "ErrFileAppError",
			expectedError: app.ErrFileAppError,
			expectedClass: errutil.ErrorClassValidation,
			expectedCode:  400,
		},
	}

	// Verify the registry contains expected entries
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			found := false
			for _, entry := range FileDataErrRegistry {
				if entry.ErrorIn == tt.expectedError {
					found = true
					assert.Equal(t, tt.expectedCode, entry.HandlePolicy.StatusCode)
					assert.Equal(t, tt.expectedClass, entry.HandlePolicy.ErrorClass)
					assert.NotEmpty(t, entry.HandlePolicy.PublicMsg)
					break
				}
			}
			assert.True(t, found, "Error %v should be in registry", tt.expectedError)
		})
	}
}

func TestFileDataErrRegistry_Structure(t *testing.T) {
	t.Parallel()

	require.NotEmpty(t, FileDataErrRegistry, "Registry should not be empty")

	for i, entry := range FileDataErrRegistry {
		t.Run("entry_"+string(rune(i)), func(t *testing.T) {
			t.Parallel()
			assert.NotNil(t, entry.ErrorIn, "Entry %d should have ErrorIn", i)
			assert.NotZero(t, entry.HandlePolicy.StatusCode, "Entry %d should have StatusCode", i)
			assert.NotEmpty(t, entry.HandlePolicy.PublicMsg, "Entry %d should have PublicMsg", i)
			assert.NotNil(t, entry.HandlePolicy.ErrorClass, "Entry %d should have ErrorClass", i)
		})
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		err            error
		name           string
		expectedStatus int
		expectMessages bool
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedStatus: 500,
			expectMessages: true,
		},
		{
			name:           "file not found error",
			err:            app.ErrFileNotFound,
			expectedStatus: 404,
			expectMessages: true,
		},
		{
			name:           "access denied error",
			err:            app.ErrFileAccessDenied,
			expectedStatus: 403,
			expectMessages: true,
		},
		{
			name:           "validation error",
			err:            app.ErrFileIncorrectStorageKey,
			expectedStatus: 400,
			expectMessages: true,
		},
		{
			name:           "unknown error",
			err:            assert.AnError,
			expectedStatus: 500,
			expectMessages: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx, _ := gin.CreateTestContext(nil)

			statusCode, messages := handleError(tt.err, ctx)

			assert.Equal(t, tt.expectedStatus, statusCode)
			if tt.expectMessages {
				assert.NotEmpty(t, messages, "Should have error messages")
			}
		})
	}
}
