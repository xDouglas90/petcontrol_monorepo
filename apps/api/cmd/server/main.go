package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/config"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/handler"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/queue"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool, err := db.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	queries := sqlc.New(pool)
	userService := service.NewUserService(queries)
	authService := service.NewAuthService(queries, cfg.JWTSecret, cfg.JWTTTL)
	workerPublisher := queue.NewAsynqPublisher(cfg.RedisAddr, cfg.WorkerQueue)
	defer func() {
		if err := workerPublisher.Close(); err != nil {
			log.Printf("close worker publisher: %v", err)
		}
	}()
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	workerHandler := handler.NewWorkerHandler(workerPublisher)
	healthHandler := handler.NewHealthHandler(pool)

	router := gin.New()
	router.Use(gin.Logger(), middleware.Recovery())

	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	v1 := router.Group("/api/v1")
	v1.POST("/auth/login", authHandler.Login)

	protected := v1.Group("/")
	protected.Use(middleware.Auth(cfg.JWTSecret), middleware.Tenant())
	protected.GET("/users", userHandler.List)
	protected.GET("/company-users", userHandler.ListCompanyUsers)
	protected.POST("/worker/notifications/dummy", workerHandler.EnqueueDummyNotification)
	protected.GET("/modules/:code/access", middleware.RequireModule(queries, ""), func(c *gin.Context) {
		c.JSON(200, gin.H{"data": gin.H{"allowed": true, "module": c.Param("code")}})
	})

	log.Printf("api listening on %s", cfg.Address())
	if err := router.Run(cfg.Address()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
