package handlers

import (
	"net/http"
	"os"

	"github.com/aioutlet/cart-service/internal/middleware"
	"github.com/aioutlet/cart-service/internal/models"
	"github.com/aioutlet/cart-service/internal/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CartHandler handles HTTP requests for cart operations
type CartHandler struct {
	cartService services.CartService
	logger      *zap.Logger
}

// NewCartHandler creates a new cart handler
func NewCartHandler(cartService services.CartService, logger *zap.Logger) *CartHandler {
	return &CartHandler{
		cartService: cartService,
		logger:      logger,
	}
}

// GetCart godoc
// @Summary Get user's cart
// @Description Get the current user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.CartResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cart [get]
func (h *CartHandler) GetCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	userIDStr := userID.(string)
	ctx := c.Request.Context()

	cart, err := h.cartService.GetCart(ctx, userIDStr)
	if err != nil {
		h.logger.Error("Failed to get cart", 
			zap.String("userID", userIDStr), 
			zap.String("correlationID", middleware.GetCorrelationID(c)),
			zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, "Failed to get cart", err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "Cart retrieved successfully", cart)
}

// AddItem godoc
// @Summary Add item to cart
// @Description Add an item to the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.AddItemRequest true "Add item request"
// @Success 200 {object} models.CartResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cart/items [post]
func (h *CartHandler) AddItem(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var request models.AddItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userIDStr := userID.(string)
	ctx := c.Request.Context()

	cart, err := h.cartService.AddItem(ctx, userIDStr, request)
	if err != nil {
		h.logger.Error("Failed to add item to cart", 
			zap.String("userID", userIDStr),
			zap.String("productID", request.ProductID),
			zap.String("correlationID", middleware.GetCorrelationID(c)),
			zap.Error(err))

		statusCode := h.getErrorStatusCode(err)
		h.respondWithError(c, statusCode, err.Error(), err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "Item added to cart successfully", cart)
}

// UpdateItem godoc
// @Summary Update item in cart
// @Description Update the quantity of an item in the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param productId path string true "Product ID"
// @Param request body models.UpdateItemRequest true "Update item request"
// @Success 200 {object} models.CartResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cart/items/{productId} [put]
func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	productID := c.Param("productId")
	if productID == "" {
		h.respondWithError(c, http.StatusBadRequest, "Product ID is required", nil)
		return
	}

	var request models.UpdateItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	cart, err := h.cartService.UpdateItem(c.Request.Context(), userID.(string), productID, request)
	if err != nil {
		h.logger.Error("Failed to update item in cart", 
			zap.String("userID", userID.(string)),
			zap.String("productID", productID),
			zap.Error(err))

		statusCode := h.getErrorStatusCode(err)
		h.respondWithError(c, statusCode, err.Error(), err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "Item updated successfully", cart)
}

// RemoveItem godoc
// @Summary Remove item from cart
// @Description Remove an item from the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param productId path string true "Product ID"
// @Success 200 {object} models.CartResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cart/items/{productId} [delete]
func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	productID := c.Param("productId")
	if productID == "" {
		h.respondWithError(c, http.StatusBadRequest, "Product ID is required", nil)
		return
	}

	cart, err := h.cartService.RemoveItem(c.Request.Context(), userID.(string), productID)
	if err != nil {
		h.logger.Error("Failed to remove item from cart", 
			zap.String("userID", userID.(string)),
			zap.String("productID", productID),
			zap.Error(err))

		statusCode := h.getErrorStatusCode(err)
		h.respondWithError(c, statusCode, err.Error(), err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "Item removed successfully", cart)
}

// ClearCart godoc
// @Summary Clear cart
// @Description Remove all items from the user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.CartResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cart [delete]
func (h *CartHandler) ClearCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	err := h.cartService.ClearCart(c.Request.Context(), userID.(string))
	if err != nil {
		h.logger.Error("Failed to clear cart", 
			zap.String("userID", userID.(string)),
			zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, "Failed to clear cart", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cart cleared successfully",
	})
}

// TransferCart godoc
// @Summary Transfer guest cart to user
// @Description Transfer items from a guest cart to the authenticated user's cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.TransferCartRequest true "Transfer cart request"
// @Success 200 {object} models.CartResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /cart/transfer [post]
func (h *CartHandler) TransferCart(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var request models.TransferCartRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	cart, err := h.cartService.TransferCart(c.Request.Context(), request.GuestID, userID.(string))
	if err != nil {
		h.logger.Error("Failed to transfer cart", 
			zap.String("guestID", request.GuestID),
			zap.String("userID", userID.(string)),
			zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, "Failed to transfer cart", err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "Cart transferred successfully", cart)
}

// Guest cart handlers (no authentication required)

// GetGuestCart godoc
// @Summary Get guest cart
// @Description Get a guest user's shopping cart
// @Tags Guest Cart
// @Accept json
// @Produce json
// @Param guestId path string true "Guest ID"
// @Success 200 {object} models.CartResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /guest/cart/{guestId} [get]
func (h *CartHandler) GetGuestCart(c *gin.Context) {
	guestID := c.Param("guestId")
	if guestID == "" {
		h.respondWithError(c, http.StatusBadRequest, "Guest ID is required", nil)
		return
	}

	cart, err := h.cartService.GetCart(c.Request.Context(), guestID)
	if err != nil {
		h.logger.Error("Failed to get guest cart", 
			zap.String("guestID", guestID), 
			zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, "Failed to get cart", err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "Cart retrieved successfully", cart)
}

// AddGuestItem godoc
// @Summary Add item to guest cart
// @Description Add an item to a guest user's shopping cart
// @Tags Guest Cart
// @Accept json
// @Produce json
// @Param guestId path string true "Guest ID"
// @Param request body models.AddItemRequest true "Add item request"
// @Success 200 {object} models.CartResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /guest/cart/{guestId}/items [post]
func (h *CartHandler) AddGuestItem(c *gin.Context) {
	guestID := c.Param("guestId")
	if guestID == "" {
		h.respondWithError(c, http.StatusBadRequest, "Guest ID is required", nil)
		return
	}

	var request models.AddItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	cart, err := h.cartService.AddItem(c.Request.Context(), guestID, request)
	if err != nil {
		h.logger.Error("Failed to add item to guest cart", 
			zap.String("guestID", guestID),
			zap.String("productID", request.ProductID),
			zap.Error(err))

		statusCode := h.getErrorStatusCode(err)
		h.respondWithError(c, statusCode, err.Error(), err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "Item added to cart successfully", cart)
}

// UpdateGuestItem godoc
// @Summary Update item in guest cart
// @Description Update the quantity of an item in a guest user's shopping cart
// @Tags Guest Cart
// @Accept json
// @Produce json
// @Param guestId path string true "Guest ID"
// @Param productId path string true "Product ID"
// @Param request body models.UpdateItemRequest true "Update item request"
// @Success 200 {object} models.CartResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /guest/cart/{guestId}/items/{productId} [put]
func (h *CartHandler) UpdateGuestItem(c *gin.Context) {
	guestID := c.Param("guestId")
	if guestID == "" {
		h.respondWithError(c, http.StatusBadRequest, "Guest ID is required", nil)
		return
	}

	productID := c.Param("productId")
	if productID == "" {
		h.respondWithError(c, http.StatusBadRequest, "Product ID is required", nil)
		return
	}

	var request models.UpdateItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	cart, err := h.cartService.UpdateItem(c.Request.Context(), guestID, productID, request)
	if err != nil {
		h.logger.Error("Failed to update item in guest cart", 
			zap.String("guestID", guestID),
			zap.String("productID", productID),
			zap.Error(err))

		statusCode := h.getErrorStatusCode(err)
		h.respondWithError(c, statusCode, err.Error(), err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "Item updated successfully", cart)
}

// RemoveGuestItem godoc
// @Summary Remove item from guest cart
// @Description Remove an item from a guest user's shopping cart
// @Tags Guest Cart
// @Accept json
// @Produce json
// @Param guestId path string true "Guest ID"
// @Param productId path string true "Product ID"
// @Success 200 {object} models.CartResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /guest/cart/{guestId}/items/{productId} [delete]
func (h *CartHandler) RemoveGuestItem(c *gin.Context) {
	guestID := c.Param("guestId")
	if guestID == "" {
		h.respondWithError(c, http.StatusBadRequest, "Guest ID is required", nil)
		return
	}

	productID := c.Param("productId")
	if productID == "" {
		h.respondWithError(c, http.StatusBadRequest, "Product ID is required", nil)
		return
	}

	cart, err := h.cartService.RemoveItem(c.Request.Context(), guestID, productID)
	if err != nil {
		h.logger.Error("Failed to remove item from guest cart", 
			zap.String("guestID", guestID),
			zap.String("productID", productID),
			zap.Error(err))

		statusCode := h.getErrorStatusCode(err)
		h.respondWithError(c, statusCode, err.Error(), err)
		return
	}

	h.respondWithSuccess(c, http.StatusOK, "Item removed successfully", cart)
}

// ClearGuestCart godoc
// @Summary Clear guest cart
// @Description Remove all items from a guest user's shopping cart
// @Tags Guest Cart
// @Accept json
// @Produce json
// @Param guestId path string true "Guest ID"
// @Success 200 {object} models.CartResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /guest/cart/{guestId} [delete]
func (h *CartHandler) ClearGuestCart(c *gin.Context) {
	guestID := c.Param("guestId")
	if guestID == "" {
		h.respondWithError(c, http.StatusBadRequest, "Guest ID is required", nil)
		return
	}

	err := h.cartService.ClearCart(c.Request.Context(), guestID)
	if err != nil {
		h.logger.Error("Failed to clear guest cart", 
			zap.String("guestID", guestID),
			zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, "Failed to clear cart", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cart cleared successfully",
	})
}

// Helper methods

func (h *CartHandler) respondWithSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, models.CartResponse{
		Success: true,
		Message: message,
		Data:    data.(*models.Cart),
	})
}

func (h *CartHandler) respondWithError(c *gin.Context, statusCode int, message string, err error) {
	response := models.ErrorResponse{
		Success: false,
		Message: message,
	}

	// Add correlation ID if available
	if correlationID := c.GetString("correlation_id"); correlationID != "" {
		response.CorrelationID = correlationID
	}

	// Environment-based error details
	isDevelopment := os.Getenv("ENVIRONMENT") == "development" || os.Getenv("GO_ENV") == "development"
	
	if err != nil {
		if isDevelopment {
			response.Error = err.Error()
		} else {
			// Don't expose internal error details in production
			switch statusCode {
			case http.StatusNotFound:
				response.Error = "Resource not found"
			case http.StatusBadRequest:
				response.Error = "Invalid request"
			case http.StatusUnauthorized:
				response.Error = "Unauthorized"
			case http.StatusForbidden:
				response.Error = "Forbidden"
			default:
				response.Error = "Internal server error"
			}
		}
	}

	c.JSON(statusCode, response)
}

func (h *CartHandler) getErrorStatusCode(err error) int {
	switch err {
	case models.ErrCartNotFound, models.ErrItemNotFound, models.ErrProductNotFound:
		return http.StatusNotFound
	case models.ErrInsufficientStock, models.ErrMaxItemsExceeded, models.ErrMaxQuantityExceeded, models.ErrInvalidQuantity:
		return http.StatusBadRequest
	case models.ErrCartExpired:
		return http.StatusGone
	default:
		return http.StatusInternalServerError
	}
}
