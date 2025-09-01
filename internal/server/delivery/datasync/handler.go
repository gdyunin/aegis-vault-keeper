package datasync

import (
	"context"
	"net/http"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/datasync"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/response"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Service defines the data synchronization application service interface.
type Service interface {
	// Pull retrieves all user data for client synchronization.
	Pull(context.Context, uuid.UUID) (*datasync.SyncPayload, error)
	// Push accepts synchronized data from client and applies changes.
	Push(context.Context, *datasync.SyncPayload) error
}

// Handler handles HTTP requests for data synchronization endpoints.
type Handler struct {
	// s is the data sync service used to process bulk operations.
	s Service
}

// NewHandler creates a new data synchronization handler with the provided service.
func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

// Pull retrieves all user data for synchronization.
// @Summary      Pull all user data
// @Description  Retrieves all user data (cards, credentials, notes, files) for synchronization
// .
// @Tags         DataSync
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} SyncPayload "User data retrieved successfully"
// @Success      204 "No data found"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/sync [get]
// .
func (h *Handler) Pull(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	userID, err := extractor.UserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	payload, err := h.s.Pull(c, userID)
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	resp := NewSyncPayloadFromApp(payload)

	if resp.isEmpty() {
		c.Data(http.StatusNoContent, "", nil)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Push synchronizes user data to the server.
// @Summary      Push user data for synchronization
// @Description  Uploads and syncs all user data (cards, credentials, notes, files)
// .
// @Tags         DataSync
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body SyncPayload true "User data to synchronize"
// @Success      204 "Data synchronized successfully"
// @Failure      400 {object} response.Error "Bad request - invalid input data"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/sync [post]
// .
func (h *Handler) Push(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	userID, err := extractor.UserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	// req holds the deserialized JSON request payload for data synchronization.
	var req SyncPayload
	err = extractor.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	if err := h.s.Push(c, req.ToApp(userID)); err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	c.Data(http.StatusNoContent, "", nil)
}
