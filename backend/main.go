package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"intan-florist-api/config"
	"intan-florist-api/internal/handlers"
	"intan-florist-api/internal/middleware"
	"intan-florist-api/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := repository.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	h := handlers.Handler{DB: db, Config: cfg}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS(cfg.AllowedOrigins))
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api := r.Group("/api/v1")
	api.GET("/products", h.Products)
	api.GET("/products/:slug", h.ProductBySlug)
	api.GET("/categories", h.Categories)
	api.GET("/pages/:slug", h.PageBySlug)
	api.GET("/banners", h.Banners)
	api.GET("/payment-methods", h.PaymentMethods)
	api.GET("/shipping-info", h.Shipping)
	api.GET("/testimonials", h.Testimonials)
	api.GET("/settings", h.Settings)
	api.POST("/contacts", h.CreateContact)
	api.POST("/auth/login", h.Login)
	api.POST("/auth/logout", h.Logout)
	api.GET("/auth/me", middleware.Auth(cfg.JWTSecret), h.Me)

	protected := api.Group("", middleware.Auth(cfg.JWTSecret))
	registerManagement(protected, h)
	management := api.Group("/management", middleware.Auth(cfg.JWTSecret))
	registerManagement(management, h)

	server := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}

func registerManagement(group *gin.RouterGroup, h handlers.Handler) {
	resources := []string{"products", "categories", "testimonials", "banners", "pages"}
	for _, resource := range resources {
		group.POST("/"+resource, middleware.Authorize(resource), h.Create(resource))
		group.PUT("/"+resource+"/:id", middleware.Authorize(resource), h.Update(resource))
		group.DELETE("/"+resource+"/:id", middleware.Authorize(resource), h.Delete(resource))
	}
	group.POST("/products/:id/images", middleware.Authorize("products"), h.AddProductImage)
	group.DELETE("/products/images/:id", middleware.Authorize("products"), h.DeleteProductImage)
	group.GET("/users", middleware.Authorize("users"), h.ManagementList("users"))
	group.POST("/users", middleware.Authorize("users"), h.CreateUser)
	group.PUT("/users/:id", middleware.Authorize("users"), h.Update("users"))
	group.DELETE("/users/:id", middleware.Authorize("users"), h.Delete("users"))
	group.GET("/pages", middleware.Authorize("pages"), h.ManagementList("pages"))
	group.GET("/contacts", middleware.Authorize("contacts"), h.ManagementList("contacts"))
	group.PUT("/settings/:key", middleware.Authorize("settings"), h.UpdateSetting)
}

func registerRestrictedCRUD(group *gin.RouterGroup, h handlers.Handler, resource, path string) {
	group.GET(path, middleware.Authorize(resource), h.ManagementList(resource))
	group.POST(path, middleware.Authorize(resource), h.Create(resource))
	group.PUT(path+"/:id", middleware.Authorize(resource), h.Update(resource))
	group.DELETE(path+"/:id", middleware.Authorize(resource), h.Delete(resource))
}
