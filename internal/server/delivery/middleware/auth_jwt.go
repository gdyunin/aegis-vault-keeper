package middleware

import (
	"net/http"
	"strings"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/consts"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthWithJWTService defines the interface for JWT token validation services.
type AuthWithJWTService interface {
	// ValidateToken validates the provided JWT token and returns the user ID.
	ValidateToken(token string) (uuid.UUID, error)
}

// AuthWithJWT creates middleware that validates JWT tokens in the Authorization header.
// It extracts the Bearer token, validates it using the provided service, and sets the user ID in context.
func AuthWithJWT(service AuthWithJWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.Request.Header.Get("Authorization")
		if accessToken == "" {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
		rawToken := strings.TrimPrefix(accessToken, "Bearer ")

		userID, err := service.ValidateToken(rawToken)
		if err != nil {
			code, msgs := handleError(err, c)
			c.JSON(code, response.Error{
				Messages: msgs,
			})
			c.Abort()
			return
		}

		c.Set(consts.CtxKeyUserID, userID)

		c.Next()
	}
}
