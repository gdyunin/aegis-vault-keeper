package errutil

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_Best(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		matches   []Policy
		wantBest  Policy
		wantFound bool
	}{
		{
			name:      "empty_matches_returns_false",
			matches:   []Policy{},
			wantBest:  Policy{},
			wantFound: false,
		},
		{
			name: "single_match_returns_that_policy",
			matches: []Policy{
				{
					StatusCode: http.StatusBadRequest,
					ErrorClass: ErrorClassValidation,
					PublicMsg:  "Validation error",
				},
			},
			wantBest: Policy{
				StatusCode: http.StatusBadRequest,
				ErrorClass: ErrorClassValidation,
				PublicMsg:  "Validation error",
			},
			wantFound: true,
		},
		{
			name: "tech_error_precedes_auth_error",
			matches: []Policy{
				{
					StatusCode: http.StatusUnauthorized,
					ErrorClass: ErrorClassAuth,
					PublicMsg:  "Auth error",
				},
				{
					StatusCode: http.StatusInternalServerError,
					ErrorClass: ErrorClassTech,
					PublicMsg:  "Tech error",
				},
			},
			wantBest: Policy{
				StatusCode: http.StatusInternalServerError,
				ErrorClass: ErrorClassTech,
				PublicMsg:  "Tech error",
			},
			wantFound: true,
		},
		{
			name: "auth_error_precedes_validation_error",
			matches: []Policy{
				{
					StatusCode: http.StatusBadRequest,
					ErrorClass: ErrorClassValidation,
					PublicMsg:  "Validation error",
				},
				{
					StatusCode: http.StatusUnauthorized,
					ErrorClass: ErrorClassAuth,
					PublicMsg:  "Auth error",
				},
			},
			wantBest: Policy{
				StatusCode: http.StatusUnauthorized,
				ErrorClass: ErrorClassAuth,
				PublicMsg:  "Auth error",
			},
			wantFound: true,
		},
		{
			name: "multiple_same_class_returns_first",
			matches: []Policy{
				{
					StatusCode: http.StatusBadRequest,
					ErrorClass: ErrorClassValidation,
					PublicMsg:  "First validation error",
				},
				{
					StatusCode: http.StatusBadRequest,
					ErrorClass: ErrorClassValidation,
					PublicMsg:  "Second validation error",
				},
			},
			wantBest: Policy{
				StatusCode: http.StatusBadRequest,
				ErrorClass: ErrorClassValidation,
				PublicMsg:  "First validation error",
			},
			wantFound: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := Registry{}
			best, found := registry.Best(tt.matches)

			assert.Equal(t, tt.wantFound, found)
			if tt.wantFound {
				assert.Equal(t, tt.wantBest.StatusCode, best.StatusCode)
				assert.Equal(t, tt.wantBest.ErrorClass, best.ErrorClass)
				assert.Equal(t, tt.wantBest.PublicMsg, best.PublicMsg)
			}
		})
	}
}

func TestRegistry_Message(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		matches []Policy
		want    []string
		best    Policy
	}{
		{
			name: "no_merge_returns_single_message",
			best: Policy{
				PublicMsg:  "Single error message",
				AllowMerge: false,
			},
			matches: []Policy{
				{PublicMsg: "Single error message", AllowMerge: false},
				{PublicMsg: "Another message", AllowMerge: true},
			},
			want: []string{"Single error message"},
		},
		{
			name: "merge_compatible_messages",
			best: Policy{
				PublicMsg:  "First message",
				AllowMerge: true,
				ErrorClass: ErrorClassValidation,
				StatusCode: http.StatusBadRequest,
			},
			matches: []Policy{
				{
					PublicMsg:  "First message",
					AllowMerge: true,
					ErrorClass: ErrorClassValidation,
					StatusCode: http.StatusBadRequest,
				},
				{
					PublicMsg:  "Second message",
					AllowMerge: true,
					ErrorClass: ErrorClassValidation,
					StatusCode: http.StatusBadRequest,
				},
				{
					PublicMsg:  "Third message",
					AllowMerge: true,
					ErrorClass: ErrorClassValidation,
					StatusCode: http.StatusBadRequest,
				},
			},
			want: []string{"First message", "Second message", "Third message"},
		},
		{
			name: "merge_deduplicates_messages",
			best: Policy{
				PublicMsg:  "Duplicate message",
				AllowMerge: true,
				ErrorClass: ErrorClassValidation,
				StatusCode: http.StatusBadRequest,
			},
			matches: []Policy{
				{
					PublicMsg:  "Duplicate message",
					AllowMerge: true,
					ErrorClass: ErrorClassValidation,
					StatusCode: http.StatusBadRequest,
				},
				{
					PublicMsg:  "Duplicate message",
					AllowMerge: true,
					ErrorClass: ErrorClassValidation,
					StatusCode: http.StatusBadRequest,
				},
				{
					PublicMsg:  "Unique message",
					AllowMerge: true,
					ErrorClass: ErrorClassValidation,
					StatusCode: http.StatusBadRequest,
				},
			},
			want: []string{"Duplicate message", "Unique message"},
		},
		{
			name: "merge_excludes_incompatible_classes",
			best: Policy{
				PublicMsg:  "Auth message",
				AllowMerge: true,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusUnauthorized,
			},
			matches: []Policy{
				{
					PublicMsg:  "Auth message",
					AllowMerge: true,
					ErrorClass: ErrorClassAuth,
					StatusCode: http.StatusUnauthorized,
				},
				{
					PublicMsg:  "Validation message",
					AllowMerge: true,
					ErrorClass: ErrorClassValidation,
					StatusCode: http.StatusBadRequest,
				},
			},
			want: []string{"Auth message"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			registry := Registry{}
			got := registry.Message(tt.best, tt.matches)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegistry_Handle(t *testing.T) {
	t.Parallel()

	registry := Registry{
		{
			ErrorIn: errTestAuth,
			HandlePolicy: Policy{
				StatusCode: http.StatusUnauthorized,
				PublicMsg:  "Authentication failed",
				LogIt:      true,
				ErrorClass: ErrorClassAuth,
			},
		},
		{
			ErrorIn: errTestValidation,
			HandlePolicy: Policy{
				StatusCode: http.StatusBadRequest,
				PublicMsg:  "Validation failed",
				LogIt:      false,
				ErrorClass: ErrorClassValidation,
			},
		},
	}

	tests := []struct {
		inputErr     error
		name         string
		wantMessages []string
		wantStatus   int
		wantLogIt    bool
	}{
		{
			name:         "nil_error_returns_defaults",
			inputErr:     nil,
			wantStatus:   http.StatusInternalServerError,
			wantMessages: []string{"Internal Server Error"},
			wantLogIt:    false,
		},
		{
			name:         "unknown_error_returns_defaults",
			inputErr:     errTestTech,
			wantStatus:   http.StatusInternalServerError,
			wantMessages: []string{"Internal Server Error"},
			wantLogIt:    true,
		},
		{
			name:         "auth_error_returns_auth_policy",
			inputErr:     errTestAuth,
			wantStatus:   http.StatusUnauthorized,
			wantMessages: []string{"Authentication failed"},
			wantLogIt:    true,
		},
		{
			name:         "validation_error_returns_validation_policy",
			inputErr:     errTestValidation,
			wantStatus:   http.StatusBadRequest,
			wantMessages: []string{"Validation failed"},
			wantLogIt:    false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			status, messages, logIt := registry.Handle(tt.inputErr)

			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantMessages, messages)
			assert.Equal(t, tt.wantLogIt, logIt)
		})
	}
}

func TestHandleWithRegistry(t *testing.T) {
	t.Parallel()

	registry := Registry{
		{
			ErrorIn: errTestAuth,
			HandlePolicy: Policy{
				StatusCode: http.StatusUnauthorized,
				PublicMsg:  "Authentication failed",
				LogIt:      true,
				ErrorClass: ErrorClassAuth,
			},
		},
		{
			ErrorIn: errTestValidation,
			HandlePolicy: Policy{
				StatusCode: http.StatusBadRequest,
				PublicMsg:  "Validation failed",
				LogIt:      false,
				ErrorClass: ErrorClassValidation,
			},
		},
	}

	tests := []struct {
		inputErr     error
		name         string
		wantMessages []string
		wantStatus   int
		expectLogged bool
	}{
		{
			name:         "auth_error_gets_logged",
			inputErr:     errTestAuth,
			wantStatus:   http.StatusUnauthorized,
			wantMessages: []string{"Authentication failed"},
			expectLogged: true,
		},
		{
			name:         "validation_error_not_logged",
			inputErr:     errTestValidation,
			wantStatus:   http.StatusBadRequest,
			wantMessages: []string{"Validation failed"},
			expectLogged: false,
		},
		{
			name:         "unknown_error_gets_logged",
			inputErr:     errTestTech,
			wantStatus:   http.StatusInternalServerError,
			wantMessages: []string{"Internal Server Error"},
			expectLogged: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(nil)

			status, messages := HandleWithRegistry(registry, tt.inputErr, c)

			assert.Equal(t, tt.wantStatus, status)
			assert.Equal(t, tt.wantMessages, messages)

			if tt.expectLogged {
				assert.NotEmpty(t, c.Errors, "Error should be logged to gin context")
			} else {
				assert.Empty(t, c.Errors, "Error should not be logged to gin context")
			}
		})
	}
}

func TestMerge(t *testing.T) {
	t.Parallel()

	registry1 := Registry{
		{
			ErrorIn: errTestAuth,
			HandlePolicy: Policy{
				StatusCode: http.StatusUnauthorized,
				PublicMsg:  "Auth failed",
			},
		},
	}

	registry2 := Registry{
		{
			ErrorIn: errTestValidation,
			HandlePolicy: Policy{
				StatusCode: http.StatusBadRequest,
				PublicMsg:  "Validation failed",
			},
		},
	}

	registry3 := Registry{
		{
			ErrorIn: errTestTech,
			HandlePolicy: Policy{
				StatusCode: http.StatusInternalServerError,
				PublicMsg:  "Tech failed",
			},
		},
	}

	tests := []struct {
		name       string
		registries []Registry
		wantCount  int
	}{
		{
			name:       "merge_empty_registries",
			registries: []Registry{},
			wantCount:  0,
		},
		{
			name:       "merge_single_registry",
			registries: []Registry{registry1},
			wantCount:  1,
		},
		{
			name:       "merge_two_registries",
			registries: []Registry{registry1, registry2},
			wantCount:  2,
		},
		{
			name:       "merge_three_registries",
			registries: []Registry{registry1, registry2, registry3},
			wantCount:  3,
		},
		{
			name:       "merge_with_empty_registry",
			registries: []Registry{registry1, {}, registry2},
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			merged := Merge(tt.registries...)
			assert.Len(t, merged, tt.wantCount)

			// Verify that all rules from input registries are present
			if tt.wantCount > 0 {
				foundErrors := make(map[error]bool)
				for _, rule := range merged {
					foundErrors[rule.ErrorIn] = true
				}

				for _, reg := range tt.registries {
					for _, rule := range reg {
						assert.True(t, foundErrors[rule.ErrorIn],
							"Error %v should be present in merged registry", rule.ErrorIn)
					}
				}
			}
		})
	}
}
