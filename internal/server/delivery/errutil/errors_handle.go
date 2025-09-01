package errutil

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorClass represents the category of an error for prioritization.
type ErrorClass int

const (
	// ErrorClassTech represents technical/infrastructure errors (highest priority).
	ErrorClassTech ErrorClass = iota
	// ErrorClassAuth represents authentication/authorization errors.
	ErrorClassAuth
	// ErrorClassValidation represents input validation errors.
	ErrorClassValidation
	// ErrorClassGeneric represents general application errors (lowest priority).
	ErrorClassGeneric
)

// Policy defines how an error should be handled in HTTP responses.
type Policy struct {
	// PublicMsg is the user-facing error message.
	PublicMsg string
	// StatusCode is the HTTP status code to return.
	StatusCode int
	// ErrorClass categorizes the error for prioritization.
	ErrorClass ErrorClass
	// LogIt indicates whether the error should be logged.
	LogIt bool
	// AllowMerge indicates whether this error can be merged with others.
	AllowMerge bool
}

// Precedes returns true if this policy has higher priority than other.
func (p Policy) Precedes(other Policy) bool {
	return p.ErrorClass < other.ErrorClass
}

// ShouldMergeWith returns true if this policy can be merged with other.
func (p Policy) ShouldMergeWith(other Policy) bool {
	return p.AllowMerge && other.AllowMerge &&
		p.ErrorClass == other.ErrorClass &&
		p.StatusCode == other.StatusCode
}

// Rule maps a specific error to its handling policy.
type Rule struct {
	// ErrorIn is the error to match against.
	ErrorIn error
	// HandlePolicy defines how to handle the error.
	HandlePolicy Policy
}

// Registry is a collection of error handling rules.
type Registry []Rule

// Match finds all policies that apply to the given error.
func (r Registry) Match(err error) []Policy {
	if err == nil {
		return nil
	}
	// out collects matching policies for the given error.
	var out []Policy
	for _, rule := range r {
		if errors.Is(err, rule.ErrorIn) {
			out = append(out, rule.HandlePolicy)
		}
	}
	return out
}

// Best selects the highest priority policy from matched policies.
func (r Registry) Best(matches []Policy) (Policy, bool) {
	if len(matches) == 0 {
		return Policy{}, false
	}
	best := matches[0]
	for _, p := range matches[1:] {
		if p.Precedes(best) {
			best = p
		}
	}
	return best, true
}

// Message builds the final error messages, merging compatible policies.
func (r Registry) Message(best Policy, matches []Policy) []string {
	if !best.AllowMerge {
		return []string{best.PublicMsg}
	}
	seen := map[string]struct{}{}
	// parts accumulates unique public messages from mergeable policies.
	var parts []string
	for _, p := range matches {
		if best.ShouldMergeWith(p) {
			if _, ok := seen[p.PublicMsg]; !ok {
				seen[p.PublicMsg] = struct{}{}
				parts = append(parts, p.PublicMsg)
			}
		}
	}
	return parts
}

// Handle processes an error using the registry and returns response details.
func (r Registry) Handle(err error) (int, []string, bool) {
	defStatus := http.StatusInternalServerError
	defMsg := http.StatusText(defStatus)
	defLog := true

	if err == nil {
		return defStatus, []string{defMsg}, false
	}

	matches := r.Match(err)
	if len(matches) == 0 {
		return defStatus, []string{defMsg}, defLog
	}

	best, _ := r.Best(matches)
	return best.StatusCode, r.Message(best, matches), best.LogIt
}

// HandleWithRegistry processes an error and optionally logs it to Gin context.
func HandleWithRegistry(r Registry, err error, c *gin.Context) (int, []string) {
	code, msgs, logIt := r.Handle(err)
	if logIt {
		_ = c.Error(err)
	}
	return code, msgs
}

// Merge combines multiple registries into a single registry.
func Merge(regs ...Registry) Registry {
	// out holds the combined registry entries from all input registries.
	var out Registry
	for _, r := range regs {
		out = append(out, r...)
	}
	return out
}
