// internal/api/handler/product_handler.go

package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/iSparshP/product-management-system/internal/domain/model"
	"github.com/iSparshP/product-management-system/internal/usecase/product"
)

type ProductHandler struct {
	usecase product.Usecase
	logger  *zap.Logger
}

func NewProductHandler(u product.Usecase, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		usecase: u,
		logger:  logger,
	}
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var input model.CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Invalid input for CreateProduct", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product, err := h.usecase.CreateProduct(c.Request.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	id := c.Param("id")
	product, err := h.usecase.GetProductByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get product by ID", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	userID := c.Query("user_id")
	minPriceStr := c.Query("min_price")
	maxPriceStr := c.Query("max_price")
	name := c.Query("name")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	filters := make(map[string]interface{})
	if minPriceStr != "" {
		minPrice, err := strconv.ParseFloat(minPriceStr, 64)
		if err == nil {
			filters["min_price"] = minPrice
		}
	}

	if maxPriceStr != "" {
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
		if err == nil {
			filters["max_price"] = maxPrice
		}
	}

	if name != "" {
		filters["name"] = name
	}

	products, err := h.usecase.GetProducts(c.Request.Context(), userID, filters)
	if err != nil {
		h.logger.Error("Failed to get products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get products"})
		return
	}

	c.JSON(http.StatusOK, products)
}
