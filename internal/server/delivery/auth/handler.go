package auth

import (
	"context"
	"net/http"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/auth"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/response"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Service defines the authentication application service interface.
type Service interface {
	// Register creates a new user account with the provided parameters.
	Register(context.Context, auth.RegisterParams) (uuid.UUID, error)
	// Login authenticates a user and returns an access token.
	Login(context.Context, auth.LoginParams) (auth.AccessToken, error)
}

// Handler handles HTTP requests for authentication endpoints.
type Handler struct {
	// s is the authentication service used to process business logic.
	s Service
}

// NewHandler creates a new authentication handler with the provided service.
func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

// Register handles user registration.
// @Summary      Register a new user
// @Description  Creates a new user account with login and password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "User registration data"
// @Success      201 {object} RegisterResponse "User created successfully"
// @Failure      400 {object} response.Error "Bad request - invalid input data"
// @Failure      409 {object} response.Error "Conflict - user already exists"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /auth/register [post]
// .
func (h *Handler) Register(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	// req holds the deserialized JSON registration request.
	var req RegisterRequest
	err := extractor.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	serviceParams := auth.RegisterParams{
		Login:    req.Login,
		Password: req.Password,
	}

	createdUserID, err := h.s.Register(c, serviceParams)
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	resp := RegisterResponse{
		ID: createdUserID,
	}

	c.JSON(http.StatusCreated, resp)
}

// Login handles user authentication.
// @Summary      Authenticate user
// @Description  Authenticates user with login and password, returns access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "User login credentials"
// @Success      200 {object} AccessToken "Authentication successful"
// @Failure      400 {object} response.Error "Bad request - invalid input data"
// @Failure      401 {object} response.Error "Unauthorized - invalid credentials"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /auth/login [post]
// .
func (h *Handler) Login(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	// req holds the deserialized JSON login request.
	var req LoginRequest
	err := extractor.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	serviceParams := auth.LoginParams{
		Login:    req.Login,
		Password: req.Password,
	}

	accessToken, err := h.s.Login(c, serviceParams)
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	resp := AccessToken{
		AccessToken: accessToken.AccessToken,
		ExpiresAt:   accessToken.ExpiresAt,
		TokenType:   accessToken.TokenType,
	}

	c.JSON(http.StatusOK, resp)
}
