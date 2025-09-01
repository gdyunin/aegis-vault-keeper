package credential

import (
	"testing"

	app "github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialErrRegistry(t *testing.T) {
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
			errorIn:        app.ErrCredentialTechError,
			expectedStatus: 500,
			expectedMsg:    "Internal Server Error",
			expectedLogIt:  true,
			expectedMerge:  false,
			expectedClass:  errutil.ErrorClassTech,
		},
		{
			name:           "access denied error",
			errorIn:        app.ErrCredentialAccessDenied,
			expectedStatus: 403,
			expectedMsg:    "Access to this credential is denied",
			expectedLogIt:  false,
			expectedMerge:  false,
			expectedClass:  errutil.ErrorClassAuth,
		},
		{
			name:           "not found error",
			errorIn:        app.ErrCredentialNotFound,
			expectedStatus: 404,
			expectedMsg:    "Credential not found",
			expectedLogIt:  false,
			expectedMerge:  false,
			expectedClass:  errutil.ErrorClassGeneric,
		},
		{
			name:           "incorrect login error",
			errorIn:        app.ErrCredentialIncorrectLogin,
			expectedStatus: 400,
			expectedMsg:    "Invalid login",
			expectedLogIt:  false,
			expectedMerge:  true,
			expectedClass:  errutil.ErrorClassValidation,
		},
		{
			name:           "incorrect password error",
			errorIn:        app.ErrCredentialIncorrectPassword,
			expectedStatus: 400,
			expectedMsg:    "Invalid password",
			expectedLogIt:  false,
			expectedMerge:  true,
			expectedClass:  errutil.ErrorClassValidation,
		},
		{
			name:           "app error",
			errorIn:        app.ErrCredentialAppError,
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
			for _, rule := range CredentialErrRegistry {
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

func TestCredentialErrRegistry_Coverage(t *testing.T) {
	t.Parallel()

	// Verify all expected credential errors are covered
	expectedErrors := []error{
		app.ErrCredentialTechError,
		app.ErrCredentialAccessDenied,
		app.ErrCredentialNotFound,
		app.ErrCredentialIncorrectLogin,
		app.ErrCredentialIncorrectPassword,
		app.ErrCredentialAppError,
	}

	registryErrors := make(map[error]bool)
	for _, rule := range CredentialErrRegistry {
		registryErrors[rule.ErrorIn] = true
	}

	for _, expectedErr := range expectedErrors {
		assert.True(t, registryErrors[expectedErr], "Error %v should be in registry", expectedErr)
	}

	// Verify registry has exactly the expected number of rules
	assert.Len(t, CredentialErrRegistry, len(expectedErrors))
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
			name:           "credential tech error",
			inputError:     app.ErrCredentialTechError,
			expectedStatus: 500,
			expectedMsgs:   []string{"Internal Server Error"},
		},
		{
			name:           "credential access denied",
			inputError:     app.ErrCredentialAccessDenied,
			expectedStatus: 403,
			expectedMsgs:   []string{"Access to this credential is denied"},
		},
		{
			name:           "credential not found",
			inputError:     app.ErrCredentialNotFound,
			expectedStatus: 404,
			expectedMsgs:   []string{"Credential not found"},
		},
		{
			name:           "incorrect login",
			inputError:     app.ErrCredentialIncorrectLogin,
			expectedStatus: 400,
			expectedMsgs:   []string{"Invalid login"},
		},
		{
			name:           "incorrect password",
			inputError:     app.ErrCredentialIncorrectPassword,
			expectedStatus: 400,
			expectedMsgs:   []string{"Invalid password"},
		},
		{
			name:           "credential app error",
			inputError:     app.ErrCredentialAppError,
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
