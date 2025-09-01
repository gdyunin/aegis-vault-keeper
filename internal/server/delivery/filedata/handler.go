package filedata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/filedata"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/response"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Service defines the file data application service interface.
type Service interface {
	// Pull retrieves a specific file by ID for the authenticated user.
	Pull(context.Context, filedata.PullParams) (*filedata.FileData, error)
	// List retrieves all file metadata belonging to the authenticated user.
	List(context.Context, filedata.ListParams) ([]*filedata.FileData, error)
	// Push uploads and stores a new file for the authenticated user.
	Push(context.Context, *filedata.PushParams) (uuid.UUID, error)
}

// Handler handles HTTP requests for file data storage endpoints.
type Handler struct {
	// s is the file data service used to process file operations.
	s Service
}

// NewHandler creates a new file data handler with the provided service.
func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

// Pull retrieves a specific file by ID.
// @Summary      Get file by ID
// @Description  Retrieves a specific file belonging to the authenticated user
// @Tags         Files
// @Accept       json
// @Produce      application/octet-stream
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "File ID" format(uuid)
// @Success      200 {file} binary "File content"
// @Failure      400 {object} response.Error "Bad request - invalid ID format"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      404 {object} response.Error "Not found - file not found"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/filedata/{id} [get]
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

	fd, err := h.s.Pull(c, filedata.PullParams{ID: pullingID, UserID: userID})
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	// buf holds the multipart form data buffer for the download response.
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	metadataWriter, err := writer.CreateFormField("metadata")
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	metadata := NewFileDataFromApp(fd).withoutData()
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	if _, err := metadataWriter.Write(metadataJSON); err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	fileWriter, err := writer.CreateFormFile("file", fd.StorageKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	if _, err := fileWriter.Write(fd.Data); err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	if err := writer.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	contentType := writer.FormDataContentType()
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=\""+fd.StorageKey+"\"")
	c.Data(http.StatusOK, contentType, buf.Bytes())
}

// List retrieves all files metadata for the authenticated user.
// @Summary      List all files
// @Description  Retrieves metadata for all files belonging to the authenticated user (without file content)
// @Tags         Files
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} ListResponse "Files metadata retrieved successfully"
// @Success      204 "No files found"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/filedata [get]
// .
func (h *Handler) List(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	userID, err := extractor.UserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	files, err := h.s.List(c, filedata.ListParams{UserID: userID})
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	if len(files) == 0 {
		c.Data(http.StatusNoContent, "", nil)
		return
	}

	c.JSON(http.StatusOK, ListResponse{Files: NewFileDataListFromApp(files)})
}

// Push uploads a new file or updates an existing one.
// @Summary      Upload or update file
// @Description  Uploads a new file or updates an existing one if ID is provided in URL path
// @Tags         Files
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        id path string false "File ID for update operation" format(uuid)
// @Param        file formData file true "File to upload"
// @Param        storage_key formData string false "Custom storage key (filename)"
// @Param        description formData string false "File description"
// @Success      201 {object} PushResponse "File uploaded successfully"
// @Failure      400 {object} response.Error "Bad request - invalid input data or file"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      404 {object} response.Error "Not found - file not found for update"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/filedata [post]
// @Router       /items/filedata/{id} [put]
// .
func (h *Handler) Push(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	userID, err := extractor.UserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	// req holds the deserialized form data for the push request.
	var req PushRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{
			Messages: []string{"File is required"},
		})
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			_ = c.Error(fmt.Errorf("failed to close uploaded file: %w", err))
		}
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error{
			Messages: []string{"Failed to read file content"},
		})
		return
	}

	fileDataID := uuid.Nil
	if idStr := c.Param("id"); idStr != "" {
		if id, err := uuid.Parse(idStr); err != nil {
			c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
			return
		} else {
			fileDataID = id
		}
	}

	newID, err := h.s.Push(c, &filedata.PushParams{
		ID:          fileDataID,
		UserID:      userID,
		StorageKey:  req.StorageKey,
		Description: req.Description,
		Data:        content,
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
