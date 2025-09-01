package credential

import (
	"context"
	"net/http"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/credential"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/response"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Service defines the credential application service interface.
type Service interface {
	// Pull retrieves a specific credential by ID for the authenticated user.
	Pull(context.Context, credential.PullParams) (*credential.Credential, error)
	// List retrieves all credentials belonging to the authenticated user.
	List(context.Context, credential.ListParams) ([]*credential.Credential, error)
	// Push creates or updates a credential for the authenticated user.
	Push(context.Context, *credential.PushParams) (uuid.UUID, error)
}

// Handler handles HTTP requests for credential endpoints.
type Handler struct {
	// s is the credential service used to process business logic.
	s Service
}

// NewHandler creates a new credential handler with the provided service.
func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

// Pull retrieves a specific credential by ID.
// @Summary      Get credential by ID
// @Description  Retrieves a specific credential belonging to the authenticated user
// @Tags         Credentials
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Credential ID" format(uuid)
// @Success      200 {object} PullResponse "Credential retrieved successfully"
// @Failure      400 {object} response.Error "Bad request - invalid ID format"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      404 {object} response.Error "Not found - credential not found"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/credentials/{id} [get]
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

	cred, err := h.s.Pull(c, credential.PullParams{ID: pullingID, UserID: userID})
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	resp := PullResponse{Credential: NewCredentialFromApp(cred)}
	c.JSON(http.StatusOK, resp)
}

// List retrieves all credentials for the authenticated user.
// @Summary      List all credentials
// @Description  Retrieves all credentials belonging to the authenticated user
// @Tags         Credentials
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} ListResponse "Credentials retrieved successfully"
// @Success      204 "No credentials found"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/credentials [get]
// .
func (h *Handler) List(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	userID, err := extractor.UserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	creds, err := h.s.List(c, credential.ListParams{UserID: userID})
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	if len(creds) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	resp := ListResponse{Credentials: NewCredentialsFromApp(creds)}
	c.JSON(http.StatusOK, resp)
}

// Push creates a new credential or updates an existing one.
// @Summary      Create or update credential
// @Description  Creates a new credential or updates an existing one if ID is provided in URL path
// @Tags         Credentials
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string false "Credential ID for update operation" format(uuid)
// @Param        request body PushRequest true "Credential data"
// @Success      201 {object} PushResponse "Credential created or updated successfully"
// @Failure      400 {object} response.Error "Bad request - invalid input data"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      404 {object} response.Error "Not found - credential not found for update"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/credentials [post]
// @Router       /items/credentials/{id} [put]
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

	credID := uuid.Nil
	if idStr := c.Param("id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err != nil {
			c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
			return
		} else {
			credID = id
		}
	}

	newID, err := h.s.Push(c, &credential.PushParams{
		ID:          credID,
		UserID:      userID,
		Login:       req.Login,
		Password:    req.Password,
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
