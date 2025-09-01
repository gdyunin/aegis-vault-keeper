package auth

import "github.com/gin-gonic/gin"

// RegisterRoutes registers authentication endpoints on the provided router group.
// Creates /auth/register and /auth/login endpoints with the specified handler.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	authGroup := r.Group("/auth")
	authGroup.POST("/register", h.Register)
	authGroup.POST("/login", h.Login)
}
