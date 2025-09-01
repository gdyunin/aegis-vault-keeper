package about

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// BuildInfoOperator provides access to application build information.
type BuildInfoOperator interface {
	// Version returns the application version string.
	Version() string
	// Date returns the build timestamp.
	Date() time.Time
	// Commit returns the git commit hash.
	Commit() string
}

// Handler handles HTTP requests for application information endpoints.
type Handler struct {
	// info provides build information for the application.
	info BuildInfoOperator
}

// NewHandler creates a new about handler with the provided build info operator.
func NewHandler(info BuildInfoOperator) *Handler {
	return &Handler{info: info}
}

// AboutInfo returns build information about the application.
// @Summary      Get application build information
// @Description  Returns version, build date, and commit hash of the application
// @Tags         System
// @Accept       json
// @Produce      json
// @Success      200 {object} BuildInfo "Application build information"
// @Router       /about [get].
// .
func (h *Handler) AboutInfo(c *gin.Context) {
	c.JSON(http.StatusOK, BuildInfo{
		Version: h.info.Version(),
		Date:    h.info.Date(),
		Commit:  h.info.Commit(),
	})
}
