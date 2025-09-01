package note

import (
	"context"
	"net/http"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/note"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/response"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Service defines the note application service interface.
type Service interface {
	// Pull retrieves a specific note by ID for the authenticated user.
	Pull(context.Context, note.PullParams) (*note.Note, error)
	// List retrieves all notes belonging to the authenticated user.
	List(context.Context, note.ListParams) ([]*note.Note, error)
	// Push creates or updates a note for the authenticated user.
	Push(context.Context, *note.PushParams) (uuid.UUID, error)
}

// Handler handles HTTP requests for note management endpoints.
type Handler struct {
	// s is the note service used to process note operations.
	s Service
}

// NewHandler creates a new note handler with the provided service.
func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

// Pull retrieves a specific note by ID.
// @Summary      Get note by ID
// @Description  Retrieves a specific note belonging to the authenticated user
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Note ID" format(uuid)
// @Success      200 {object} PullResponse "Note retrieved successfully"
// @Failure      400 {object} response.Error "Bad request - invalid ID format"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      404 {object} response.Error "Not found - note not found"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/notes/{id} [get]
// .
func (h *Handler) Pull(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	userID, err := extractor.UserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	// req holds the deserialized URI parameters for the pull request.
	var req PullRequest
	if err := extractor.BindURI(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	pullingID, err := uuid.Parse(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	n, err := h.s.Pull(c, note.PullParams{ID: pullingID, UserID: userID})
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	c.JSON(http.StatusOK, PullResponse{Note: NewNoteFromApp(n)})
}

// List retrieves all notes for the authenticated user.
// @Summary      List all notes
// @Description  Retrieves all notes belonging to the authenticated user
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} ListResponse "Notes retrieved successfully"
// @Success      204 "No notes found"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/notes [get]
// .
func (h *Handler) List(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	userID, err := extractor.UserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	notes, err := h.s.List(c, note.ListParams{UserID: userID})
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	if len(notes) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, ListResponse{Notes: NewNotesFromApp(notes)})
}

// Push creates a new note or updates an existing one.
// @Summary      Create or update note
// @Description  Creates a new note or updates an existing one if ID is provided in URL path
// @Tags         Notes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string false "Note ID for update operation" format(uuid)
// @Param        request body PushRequest true "Note data"
// @Success      201 {object} PushResponse "Note created or updated successfully"
// @Failure      400 {object} response.Error "Bad request - invalid input data"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      404 {object} response.Error "Not found - note not found for update"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/notes [post]
// @Router       /items/notes/{id} [put]
// .
func (h *Handler) Push(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	userID, err := extractor.UserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	// req holds the deserialized JSON request payload for the push operation.
	var req PushRequest
	if err := extractor.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	noteID := uuid.Nil
	if idStr := c.Param("id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err != nil {
			c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
			return
		} else {
			noteID = id
		}
	}

	newID, err := h.s.Push(c, &note.PushParams{
		ID:          noteID,
		UserID:      userID,
		Note:        req.Note,
		Description: req.Description,
	})
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	c.JSON(http.StatusCreated, PushResponse{ID: newID})
}
