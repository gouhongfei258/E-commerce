package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/storm/myidea/bff/config"
	"github.com/storm/myidea/bff/handler"
	"github.com/storm/myidea/bff/middleware"
)

func main() {
	cfg := config.LoadConfig()

	// --- Initialize gRPC clients ---
	orderHandler, err := handler.NewOrderHandler(cfg.GRPC.OrderServiceAddr, cfg.GRPC.DialTimeout)
	if err != nil {
		log.Fatalf("failed to dial order service: %v", err)
	}
	defer orderHandler.Close()

	productHandler, err := handler.NewProductHandler(cfg.GRPC.ProductServiceAddr, cfg.GRPC.DialTimeout)
	if err != nil {
		log.Fatalf("failed to dial product service: %v", err)
	}
	defer productHandler.Close()

	paymentHandler, err := handler.NewPaymentHandler(cfg.GRPC.PaymentServiceAddr, cfg.GRPC.DialTimeout)
	if err != nil {
		log.Fatalf("failed to dial payment service: %v", err)
	}
	defer paymentHandler.Close()

	cartHandler, err := handler.NewCartHandler(cfg.GRPC.CartServiceAddr, cfg.GRPC.DialTimeout,
		orderHandler.OrderServiceClient(),
		paymentHandler.PaymentServiceClient(),
	)
	if err != nil {
		log.Fatalf("failed to dial cart service: %v", err)
	}
	defer cartHandler.Close()

	userHandler, err := handler.NewUserHandler(cfg.GRPC.UserServiceAddr, cfg.GRPC.DialTimeout, cfg.JWT.Secret, cfg.JWT.Expiration)
	if err != nil {
		log.Fatalf("failed to dial user service: %v", err)
	}
	defer userHandler.Close()

	// --- Gin engine ---
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Trace())

	// Health check (no auth).
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Register routes.
	registerOrderRoutes(r, orderHandler, cfg.JWT.Secret)
	registerProductRoutes(r, productHandler, cfg.JWT.Secret)
	registerCartRoutes(r, cartHandler, cfg.JWT.Secret)
	registerUserRoutes(r, userHandler, cfg.JWT.Secret)
	registerPaymentRoutes(r, paymentHandler, cfg.JWT.Secret)

	// Admin routes.
	adminHandler := handler.NewAdminHandler(
		orderHandler.OrderServiceClient(),
		productHandler.ProductServiceClient(),
		userHandler.UserServiceClient(),
		paymentHandler.PaymentServiceClient(),
	)
	registerAdminRoutes(r, adminHandler, cfg.JWT.Secret)

	// --- HTTP server ---
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// --- Graceful shutdown ---
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	log.Printf("BFF listening on %s", cfg.Server.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server: %v", err)
	}
}

// registerOrderRoutes registers order-related API routes.
func registerOrderRoutes(r *gin.Engine, h *handler.OrderHandler, jwtSecret string) {
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		api.POST("/orders", h.CreateOrder)
		api.GET("/orders", h.ListOrders)
		api.GET("/orders/:id", h.GetOrder)
		api.PUT("/orders/:id/status", h.UpdateOrderStatus)
		api.POST("/orders/:id/cancel", h.CancelOrder)
	}
}

// registerCartRoutes registers cart-related API routes.
func registerCartRoutes(r *gin.Engine, h *handler.CartHandler, jwtSecret string) {
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		api.GET("/cart", h.ListItems)
		api.POST("/cart/items", h.AddItem)
		api.PUT("/cart/items/:id", h.UpdateQuantity)
		api.DELETE("/cart/items/:id", h.RemoveItem)
		api.POST("/cart/checkout", h.Checkout)
	}
}

// registerProductRoutes registers product-related API routes.
func registerProductRoutes(r *gin.Engine, h *handler.ProductHandler, jwtSecret string) {
	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		// Category
		api.GET("/categories", h.GetCategoryTree)
		api.POST("/categories", h.CreateCategory)
		api.PUT("/categories/:id", h.UpdateCategory)
		api.DELETE("/categories/:id", h.DeleteCategory)

		// Brand
		api.GET("/brands", h.ListBrands)
		api.POST("/brands", h.CreateBrand)

		// SPU
		api.GET("/spus", h.ListSPUs)
		api.GET("/spus/:id", h.GetSPU)
		api.POST("/spus", h.CreateSPU)
		api.PUT("/spus/:id", h.UpdateSPU)
		api.PUT("/spus/:id/status", h.UpdateSPUStatus)

		// SKU
		api.GET("/skus", h.ListSKUs)
		api.POST("/skus/batch", h.BatchCreateSKU)
		api.PUT("/skus/:id", h.UpdateSKU)
	}
}

// registerUserRoutes registers auth and address routes.
func registerUserRoutes(r *gin.Engine, h *handler.UserHandler, jwtSecret string) {
	// Auth routes — no JWT required.
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
	}

	// Address routes — JWT required.
	addr := r.Group("/api/v1/addresses")
	addr.Use(middleware.JWTAuth(jwtSecret))
	{
		addr.GET("", h.ListAddresses)
		addr.POST("", h.CreateAddress)
		addr.PUT("/:id", h.UpdateAddress)
		addr.DELETE("/:id", h.DeleteAddress)
		addr.PUT("/:id/default", h.SetDefaultAddress)
	}
}

// registerPaymentRoutes registers payment-related API routes.
func registerPaymentRoutes(r *gin.Engine, h *handler.PaymentHandler, jwtSecret string) {
	api := r.Group("/api/v1/payments")
	api.Use(middleware.JWTAuth(jwtSecret))
	{
		api.POST("", h.CreatePayment)
		api.POST("/:id/process", h.ProcessPayment)
		api.GET("/:id", h.GetPayment)
		api.GET("/by-order/:orderNo", h.GetPaymentByOrder)
	}

	// Notify endpoint simulates third-party callback — no JWT required
	// (in production, each provider's webhook would have its own auth mechanism).
	r.POST("/api/v1/payments/:id/notify", h.NotifyPayment)
}

// registerAdminRoutes registers admin management API routes.
func registerAdminRoutes(r *gin.Engine, h *handler.AdminHandler, jwtSecret string) {
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.JWTAuth(jwtSecret), middleware.AdminAuth())
	{
		admin.GET("/dashboard", h.Dashboard)

		admin.GET("/orders", h.ListOrders)
		admin.GET("/orders/:id", h.GetOrder)
		admin.POST("/orders/:id/ship", h.ShipOrder)
		admin.POST("/orders/:id/refund", h.RefundOrder)

		admin.GET("/categories", h.GetCategoryTree)
		admin.POST("/categories", h.CreateCategory)
		admin.PUT("/categories/:id", h.UpdateCategory)
		admin.DELETE("/categories/:id", h.DeleteCategory)

		admin.GET("/brands", h.ListBrands)
		admin.POST("/brands", h.CreateBrand)

		admin.GET("/spus", h.ListSPUs)
		admin.POST("/spus", h.CreateSPU)
		admin.PUT("/spus/:id", h.UpdateSPU)
		admin.GET("/spus/:id/skus", h.ListSKUs)

		admin.POST("/skus", h.BatchCreateSKU)
		admin.PUT("/skus/:id", h.UpdateSKU)

		admin.GET("/users", h.ListUsers)
		admin.GET("/payments", h.ListPayments)
	}
}
