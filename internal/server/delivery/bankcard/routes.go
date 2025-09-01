package bankcard

import "github.com/gin-gonic/gin"

// RegisterRoutes configures bank card endpoints in the router group.
// Sets up CRUD operations: POST/GET for collections, GET/PUT for individual items.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	bankcardsGroup := r.Group("/bankcards")
	bankcardsGroup.POST("", h.Push)
	bankcardsGroup.GET("", h.List)

	bankcardsIDGroup := bankcardsGroup.Group("/:id")
	bankcardsIDGroup.GET("", h.Pull)
	bankcardsIDGroup.PUT("", h.Push)
}
