package errutil

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test errors for testing purposes.
var (
	errTestAuth       = errors.New("auth error")
	errTestValidation = errors.New("validation error")
	errTestTech       = errors.New("technical error")
)

func TestErrorClass(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		class   ErrorClass
		wantInt int
	}{
		{
			name:    "tech_error_class",
			class:   ErrorClassTech,
			wantInt: 0,
		},
		{
			name:    "auth_error_class",
			class:   ErrorClassAuth,
			wantInt: 1,
		},
		{
			name:    "validation_error_class",
			class:   ErrorClassValidation,
			wantInt: 2,
		},
		{
			name:    "generic_error_class",
			class:   ErrorClassGeneric,
			wantInt: 3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantInt, int(tt.class))
		})
	}
}

func TestPolicy_Precedes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		policy Policy
		other  Policy
		want   bool
	}{
		{
			name: "tech_precedes_auth",
			policy: Policy{
				ErrorClass: ErrorClassTech,
			},
			other: Policy{
				ErrorClass: ErrorClassAuth,
			},
			want: true,
		},
		{
			name: "auth_precedes_validation",
			policy: Policy{
				ErrorClass: ErrorClassAuth,
			},
			other: Policy{
				ErrorClass: ErrorClassValidation,
			},
			want: true,
		},
		{
			name: "validation_precedes_generic",
			policy: Policy{
				ErrorClass: ErrorClassValidation,
			},
			other: Policy{
				ErrorClass: ErrorClassGeneric,
			},
			want: true,
		},
		{
			name: "auth_does_not_precede_tech",
			policy: Policy{
				ErrorClass: ErrorClassAuth,
			},
			other: Policy{
				ErrorClass: ErrorClassTech,
			},
			want: false,
		},
		{
			name: "same_class_does_not_precede",
			policy: Policy{
				ErrorClass: ErrorClassAuth,
			},
			other: Policy{
				ErrorClass: ErrorClassAuth,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.policy.Precedes(tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPolicy_ShouldMergeWith(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		policy Policy
		other  Policy
		want   bool
	}{
		{
			name: "both_allow_merge_same_class_same_status",
			policy: Policy{
				AllowMerge: true,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusUnauthorized,
			},
			other: Policy{
				AllowMerge: true,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusUnauthorized,
			},
			want: true,
		},
		{
			name: "first_does_not_allow_merge",
			policy: Policy{
				AllowMerge: false,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusUnauthorized,
			},
			other: Policy{
				AllowMerge: true,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusUnauthorized,
			},
			want: false,
		},
		{
			name: "second_does_not_allow_merge",
			policy: Policy{
				AllowMerge: true,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusUnauthorized,
			},
			other: Policy{
				AllowMerge: false,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusUnauthorized,
			},
			want: false,
		},
		{
			name: "different_error_class",
			policy: Policy{
				AllowMerge: true,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusUnauthorized,
			},
			other: Policy{
				AllowMerge: true,
				ErrorClass: ErrorClassValidation,
				StatusCode: http.StatusUnauthorized,
			},
			want: false,
		},
		{
			name: "different_status_code",
			policy: Policy{
				AllowMerge: true,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusUnauthorized,
			},
			other: Policy{
				AllowMerge: true,
				ErrorClass: ErrorClassAuth,
				StatusCode: http.StatusBadRequest,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.policy.ShouldMergeWith(tt.other)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRule(t *testing.T) {
	t.Parallel()

	// Test that Rule struct fields are accessible
	rule := Rule{
		ErrorIn: errTestAuth,
		HandlePolicy: Policy{
			StatusCode: http.StatusUnauthorized,
			PublicMsg:  "Authentication failed",
			LogIt:      true,
			AllowMerge: false,
			ErrorClass: ErrorClassAuth,
		},
	}

	assert.Equal(t, errTestAuth, rule.ErrorIn)
	assert.Equal(t, http.StatusUnauthorized, rule.HandlePolicy.StatusCode)
	assert.Equal(t, "Authentication failed", rule.HandlePolicy.PublicMsg)
	assert.True(t, rule.HandlePolicy.LogIt)
	assert.False(t, rule.HandlePolicy.AllowMerge)
	assert.Equal(t, ErrorClassAuth, rule.HandlePolicy.ErrorClass)
}

func TestRegistry_Match(t *testing.T) {
	t.Parallel()

	registry := Registry{
		{
			ErrorIn: errTestAuth,
			HandlePolicy: Policy{
				StatusCode: http.StatusUnauthorized,
				PublicMsg:  "Auth failed",
				ErrorClass: ErrorClassAuth,
			},
		},
		{
			ErrorIn: errTestValidation,
			HandlePolicy: Policy{
				StatusCode: http.StatusBadRequest,
				PublicMsg:  "Validation failed",
				ErrorClass: ErrorClassValidation,
			},
		},
	}

	tests := []struct {
		inputErr  error
		name      string
		wantFirst Policy
		wantCount int
	}{
		{
			name:      "matches_auth_error",
			inputErr:  errTestAuth,
			wantCount: 1,
			wantFirst: Policy{
				StatusCode: http.StatusUnauthorized,
				PublicMsg:  "Auth failed",
				ErrorClass: ErrorClassAuth,
			},
		},
		{
			name:      "matches_validation_error",
			inputErr:  errTestValidation,
			wantCount: 1,
			wantFirst: Policy{
				StatusCode: http.StatusBadRequest,
				PublicMsg:  "Validation failed",
				ErrorClass: ErrorClassValidation,
			},
		},
		{
			name:      "no_match_for_unknown_error",
			inputErr:  errTestTech,
			wantCount: 0,
		},
		{
			name:      "nil_error_returns_empty",
			inputErr:  nil,
			wantCount: 0,
		},
		{
			name:      "wrapped_error_matches",
			inputErr:  errors.Join(errors.New("wrapper"), errTestAuth),
			wantCount: 1,
			wantFirst: Policy{
				StatusCode: http.StatusUnauthorized,
				PublicMsg:  "Auth failed",
				ErrorClass: ErrorClassAuth,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			matches := registry.Match(tt.inputErr)

			assert.Len(t, matches, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantFirst.StatusCode, matches[0].StatusCode)
				assert.Equal(t, tt.wantFirst.PublicMsg, matches[0].PublicMsg)
				assert.Equal(t, tt.wantFirst.ErrorClass, matches[0].ErrorClass)
			}
		})
	}
}
