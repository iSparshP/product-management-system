// internal/api/router/router.go

package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/iSparshP/product-management-system/internal/api/handler"
	"github.com/iSparshP/product-management-system/internal/api/middleware"
)

// SetupRouter initializes the Gin router with necessary middleware and routes.
func SetupRouter(productHandler *handler.ProductHandler, logger *zap.Logger) *gin.Engine {
	r := gin.New()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	r.Use(gin.Recovery())
	r.Use(middleware.LoggingMiddleware(logger))

	v1 := r.Group("/api/v1")
	{
		products := v1.Group("/products")
		{
			products.POST("", productHandler.CreateProduct)
			products.GET("/:id", productHandler.GetProductByID)
			products.GET("", productHandler.GetProducts)
		}
	}

	// Health Check Endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	return r
}
