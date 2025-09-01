package about

import "github.com/gin-gonic/gin"

// RegisterRoutes registers application information routes with the provided router group.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/about", h.AboutInfo)
}
