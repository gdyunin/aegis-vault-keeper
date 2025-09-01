package filedata

import "github.com/gin-gonic/gin"

// RegisterRoutes registers file data management routes with the provided router.
func RegisterRoutes(r gin.IRouter, h *Handler) {
	filedata := r.Group("/filedata")
	{
		filedata.GET("/:id", h.Pull)
		filedata.GET("/", h.List)
		filedata.POST("/", h.Push)
		filedata.PUT("/:id", h.Push)
	}
}
