package bankcard

import (
	"context"
	"net/http"

	"github.com/gdyunin/aegis-vault-keeper/internal/server/application/bankcard"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/response"
	"github.com/gdyunin/aegis-vault-keeper/internal/server/delivery/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Service defines the bank card application service interface.
type Service interface {
	// Pull retrieves a specific bank card by ID for the authenticated user.
	Pull(context.Context, bankcard.PullParams) (*bankcard.BankCard, error)
	// List retrieves all bank cards belonging to the authenticated user.
	List(context.Context, bankcard.ListParams) ([]*bankcard.BankCard, error)
	// Push creates or updates a bank card for the authenticated user.
	Push(context.Context, *bankcard.PushParams) (uuid.UUID, error)
}

// Handler handles HTTP requests for bank card endpoints.
type Handler struct {
	// s is the bank card service used to process business logic.
	s Service
}

// NewHandler creates a new bank card handler with the provided service.
func NewHandler(s Service) *Handler {
	return &Handler{s: s}
}

// Pull retrieves a specific bank card by ID.
// @Summary      Get bank card by ID
// @Description  Retrieves a specific bank card belonging to the authenticated user
// @Tags         BankCards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Bank card ID" format(uuid)
// @Success      200 {object} PullResponse "Bank card retrieved successfully"
// @Failure      400 {object} response.Error "Bad request - invalid ID format"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      404 {object} response.Error "Not found - bank card not found"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/bankcards/{id} [get]
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
	err = extractor.BindURI(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	pullingID, err := uuid.Parse(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	serviceParams := bankcard.PullParams{
		ID:     pullingID,
		UserID: userID,
	}

	bc, err := h.s.Pull(c, serviceParams)
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	resp := PullResponse{
		BankCard: NewBankCardFromApp(bc),
	}

	c.JSON(http.StatusOK, resp)
}

// List retrieves all bank cards for the authenticated user.
// @Summary      List all bank cards
// @Description  Retrieves all bank cards belonging to the authenticated user
// @Tags         BankCards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} ListResponse "Bank cards retrieved successfully"
// @Success      204 "No bank cards found"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/bankcards [get]
// .
func (h *Handler) List(c *gin.Context) {
	extractor := util.NewCtxExtractor(c)

	userID, err := extractor.UserID()
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.DefaultInternalServerError)
		return
	}

	serviceParams := bankcard.ListParams{
		UserID: userID,
	}

	bcs, err := h.s.List(c, serviceParams)
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	if len(bcs) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	resp := ListResponse{
		BankCards: NewBankCardsFromApp(bcs),
	}

	c.JSON(http.StatusOK, resp)
}

// Push creates a new bank card or updates an existing one.
// @Summary      Create or update bank card
// @Description  Creates a new bank card or updates an existing one if ID is provided in URL path
// @Tags         BankCards
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string false "Bank card ID for update operation" format(uuid)
// @Param        request body PushRequest true "Bank card data"
// @Success      201 {object} PushResponse "Bank card created or updated successfully"
// @Failure      400 {object} response.Error "Bad request - invalid input data"
// @Failure      401 {object} response.Error "Unauthorized - invalid or missing token"
// @Failure      404 {object} response.Error "Not found - bank card not found for update"
// @Failure      500 {object} response.Error "Internal server error"
// @Router       /items/bankcards [post]
// @Router       /items/bankcards/{id} [put]
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
	err = extractor.BindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
		return
	}

	// cardID holds the parsed card identifier from the URL parameter.
	var cardID = uuid.Nil
	idStr := c.Param("id")
	if idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.DefaultBadRequestError)
			return
		}
		cardID = id
	}

	serviceParams := bankcard.PushParams{
		ID:          cardID,
		UserID:      userID,
		CardNumber:  req.CardNumber,
		CardHolder:  req.CardHolder,
		ExpiryMonth: req.ExpiryMonth,
		ExpiryYear:  req.ExpiryYear,
		CVV:         req.CVV,
		Description: req.Description,
	}

	createdBankCardID, err := h.s.Push(c, &serviceParams)
	if err != nil {
		code, msgs := handleError(err, c)
		c.JSON(code, response.Error{
			Messages: msgs,
		})
		return
	}

	resp := PushResponse{
		ID: createdBankCardID,
	}

	c.JSON(http.StatusCreated, resp)
}
