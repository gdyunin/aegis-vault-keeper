package note

import "github.com/gin-gonic/gin"

// RegisterRoutes registers note management routes with the provided router group.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	notesGroup := r.Group("/notes")
	notesGroup.POST("", h.Push)
	notesGroup.GET("", h.List)

	notesIDGroup := notesGroup.Group("/:id")
	notesIDGroup.GET("", h.Pull)
	notesIDGroup.PUT("", h.Push)
}
