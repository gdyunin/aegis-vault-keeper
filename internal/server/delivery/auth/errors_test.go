package auth

import (
	"errors"
	"testing"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/auth"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/errutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthErrRegistry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		errorIn        error
		expectedPolicy errutil.Policy
		found          bool
	}{
		{
			name:    "tech error",
			errorIn: auth.ErrAuthTechError,
			expectedPolicy: errutil.Policy{
				StatusCode: 500,
				PublicMsg:  "Internal Server Error",
				LogIt:      true,
				AllowMerge: false,
				ErrorClass: errutil.ErrorClassTech,
			},
			found: true,
		},
		{
			name:    "wrong login or password",
			errorIn: auth.ErrAuthWrongLoginOrPassword,
			expectedPolicy: errutil.Policy{
				StatusCode: 401,
				PublicMsg:  "The provided login or password is incorrect",
				LogIt:      false,
				AllowMerge: false,
				ErrorClass: errutil.ErrorClassAuth,
			},
			found: true,
		},
		{
			name:    "invalid access token",
			errorIn: auth.ErrAuthInvalidAccessToken,
			expectedPolicy: errutil.Policy{
				StatusCode: 401,
				PublicMsg:  "Your access token is invalid or has expired. Please log in",
				LogIt:      false,
				AllowMerge: false,
				ErrorClass: errutil.ErrorClassAuth,
			},
			found: true,
		},
		{
			name:    "incorrect login",
			errorIn: auth.ErrAuthIncorrectLogin,
			expectedPolicy: errutil.Policy{
				StatusCode: 400,
				PublicMsg:  "The login provided is not valid",
				LogIt:      false,
				AllowMerge: true,
				ErrorClass: errutil.ErrorClassValidation,
			},
			found: true,
		},
		{
			name:    "incorrect password",
			errorIn: auth.ErrAuthIncorrectPassword,
			expectedPolicy: errutil.Policy{
				StatusCode: 400,
				PublicMsg:  "The password provided is not valid",
				LogIt:      false,
				AllowMerge: true,
				ErrorClass: errutil.ErrorClassValidation,
			},
			found: true,
		},
		{
			name:    "user already exists",
			errorIn: auth.ErrAuthUserAlreadyExists,
			expectedPolicy: errutil.Policy{
				StatusCode: 409,
				PublicMsg:  "User with this login already exists",
				LogIt:      false,
				AllowMerge: false,
				ErrorClass: errutil.ErrorClassValidation,
			},
			found: true,
		},
		{
			name:    "app error",
			errorIn: auth.ErrAuthAppError,
			expectedPolicy: errutil.Policy{
				StatusCode: 400,
				PublicMsg:  "The parameters provided are invalid",
				LogIt:      false,
				AllowMerge: false,
				ErrorClass: errutil.ErrorClassValidation,
			},
			found: true,
		},
		{
			name:    "unknown error",
			errorIn: errors.New("unknown error"),
			found:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Find the error in registry
			var found bool
			var actualPolicy errutil.Policy

			for _, entry := range AuthErrRegistry {
				if errors.Is(tt.errorIn, entry.ErrorIn) {
					found = true
					actualPolicy = entry.HandlePolicy
					break
				}
			}

			assert.Equal(t, tt.found, found, "Error should be found in registry: %v", tt.found)

			if tt.found {
				assert.Equal(t, tt.expectedPolicy.StatusCode, actualPolicy.StatusCode)
				assert.Equal(t, tt.expectedPolicy.PublicMsg, actualPolicy.PublicMsg)
				assert.Equal(t, tt.expectedPolicy.LogIt, actualPolicy.LogIt)
				assert.Equal(t, tt.expectedPolicy.AllowMerge, actualPolicy.AllowMerge)
				assert.Equal(t, tt.expectedPolicy.ErrorClass, actualPolicy.ErrorClass)
			}
		})
	}
}

func TestAuthErrRegistry_AllErrorsCovered(t *testing.T) {
	t.Parallel()

	// List of all auth errors that should be in the registry
	expectedErrors := []error{
		auth.ErrAuthTechError,
		auth.ErrAuthWrongLoginOrPassword,
		auth.ErrAuthInvalidAccessToken,
		auth.ErrAuthIncorrectLogin,
		auth.ErrAuthIncorrectPassword,
		auth.ErrAuthUserAlreadyExists,
		auth.ErrAuthAppError,
	}

	// Check that all expected errors are in the registry
	for _, expectedErr := range expectedErrors {
		found := false
		for _, entry := range AuthErrRegistry {
			if errors.Is(expectedErr, entry.ErrorIn) {
				found = true
				break
			}
		}
		assert.True(t, found, "Error %v should be in the registry", expectedErr)
	}

	// Verify registry has expected number of entries
	assert.Len(
		t,
		AuthErrRegistry,
		len(expectedErrors),
		"Registry should contain exactly %d entries",
		len(expectedErrors),
	)
}

func TestAuthErrRegistry_StatusCodes(t *testing.T) {
	t.Parallel()

	statusCodeTests := []struct {
		error      error
		statusCode int
	}{
		{auth.ErrAuthTechError, 500},
		{auth.ErrAuthWrongLoginOrPassword, 401},
		{auth.ErrAuthInvalidAccessToken, 401},
		{auth.ErrAuthIncorrectLogin, 400},
		{auth.ErrAuthIncorrectPassword, 400},
		{auth.ErrAuthUserAlreadyExists, 409},
		{auth.ErrAuthAppError, 400},
	}

	for _, tt := range statusCodeTests {
		t.Run(tt.error.Error(), func(t *testing.T) {
			t.Parallel()

			for _, entry := range AuthErrRegistry {
				if errors.Is(tt.error, entry.ErrorIn) {
					assert.Equal(t, tt.statusCode, entry.HandlePolicy.StatusCode)
					return
				}
			}
			t.Errorf("Error %v not found in registry", tt.error)
		})
	}
}

func TestAuthErrRegistry_ErrorClasses(t *testing.T) {
	t.Parallel()

	errorClassTests := []struct {
		error      error
		errorClass errutil.ErrorClass
	}{
		{auth.ErrAuthTechError, errutil.ErrorClassTech},
		{auth.ErrAuthWrongLoginOrPassword, errutil.ErrorClassAuth},
		{auth.ErrAuthInvalidAccessToken, errutil.ErrorClassAuth},
		{auth.ErrAuthIncorrectLogin, errutil.ErrorClassValidation},
		{auth.ErrAuthIncorrectPassword, errutil.ErrorClassValidation},
		{auth.ErrAuthUserAlreadyExists, errutil.ErrorClassValidation},
		{auth.ErrAuthAppError, errutil.ErrorClassValidation},
	}

	for _, tt := range errorClassTests {
		t.Run(tt.error.Error(), func(t *testing.T) {
			t.Parallel()

			for _, entry := range AuthErrRegistry {
				if errors.Is(tt.error, entry.ErrorIn) {
					assert.Equal(t, tt.errorClass, entry.HandlePolicy.ErrorClass)
					return
				}
			}
			t.Errorf("Error %v not found in registry", tt.error)
		})
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputError         error
		name               string
		expectedMessages   []string
		expectedStatusCode int
	}{
		{
			name:               "tech error",
			inputError:         auth.ErrAuthTechError,
			expectedStatusCode: 500,
			expectedMessages:   []string{"Internal Server Error"},
		},
		{
			name:               "wrong login or password",
			inputError:         auth.ErrAuthWrongLoginOrPassword,
			expectedStatusCode: 401,
			expectedMessages:   []string{"The provided login or password is incorrect"},
		},
		{
			name:               "invalid access token",
			inputError:         auth.ErrAuthInvalidAccessToken,
			expectedStatusCode: 401,
			expectedMessages:   []string{"Your access token is invalid or has expired. Please log in"},
		},
		{
			name:               "incorrect login",
			inputError:         auth.ErrAuthIncorrectLogin,
			expectedStatusCode: 400,
			expectedMessages:   []string{"The login provided is not valid"},
		},
		{
			name:               "incorrect password",
			inputError:         auth.ErrAuthIncorrectPassword,
			expectedStatusCode: 400,
			expectedMessages:   []string{"The password provided is not valid"},
		},
		{
			name:               "user already exists",
			inputError:         auth.ErrAuthUserAlreadyExists,
			expectedStatusCode: 409,
			expectedMessages:   []string{"User with this login already exists"},
		},
		{
			name:               "app error",
			inputError:         auth.ErrAuthAppError,
			expectedStatusCode: 400,
			expectedMessages:   []string{"The parameters provided are invalid"},
		},
		{
			name:               "unknown error",
			inputError:         errors.New("unknown error"),
			expectedStatusCode: 500,
			expectedMessages:   []string{"Internal Server Error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup gin context
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(nil)

			// Execute
			statusCode, messages := handleError(tt.inputError, c)

			// Assert
			assert.Equal(t, tt.expectedStatusCode, statusCode)
			assert.Equal(t, tt.expectedMessages, messages)
		})
	}
}

func TestHandleError_WithWrappedErrors(t *testing.T) {
	t.Parallel()

	// Test with wrapped auth errors
	wrappedTechError := errors.New("database connection failed: " + auth.ErrAuthTechError.Error())

	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)

	statusCode, messages := handleError(wrappedTechError, c)

	// Should still handle the error as unknown since it's not directly wrapped with errors.Wrap
	assert.Equal(t, 500, statusCode)
	assert.Equal(t, []string{"Internal Server Error"}, messages)
}
