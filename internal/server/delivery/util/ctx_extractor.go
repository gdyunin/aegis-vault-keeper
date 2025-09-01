package util

import (
	"errors"
	"fmt"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/consts"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CtxExtractor provides utility methods for extracting data from Gin HTTP context.
type CtxExtractor struct {
	// c is the underlying Gin context for the current request.
	c *gin.Context
}

// NewCtxExtractor creates a new context extractor for the provided Gin context.
func NewCtxExtractor(c *gin.Context) *CtxExtractor {
	return &CtxExtractor{c: c}
}

// UserID extracts the authenticated user's ID from the context.
// Returns an error if the user ID is not present or has an invalid type.
func (e *CtxExtractor) UserID() (uuid.UUID, error) {
	value, exists := e.c.Get(consts.CtxKeyUserID)
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	userID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID in context is not a uuid.UUID")
	}
	return userID, nil
}

// BindJSON binds the request JSON body to the provided destination pointer.
// Returns an error if the JSON is malformed or doesn't match the destination type.
func (e *CtxExtractor) BindJSON(destPtr any) error {
	if err := e.c.ShouldBindJSON(destPtr); err != nil {
		return fmt.Errorf("failed to bind JSON: %w", err)
	}
	return nil
}

// BindURI binds the request URI parameters to the provided destination pointer.
// Returns an error if the URI parameters don't match the destination type.
func (e *CtxExtractor) BindURI(destPtr any) error {
	if err := e.c.ShouldBindUri(destPtr); err != nil {
		return fmt.Errorf("failed to bind URI: %w", err)
	}
	return nil
}
