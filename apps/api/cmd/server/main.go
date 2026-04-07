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
	userHandler := handler.NewUserHandler(userService)
	healthHandler := handler.NewHealthHandler(pool)

	router := gin.New()
	router.Use(gin.Logger(), middleware.Recovery())

	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	v1 := router.Group("/api/v1")
	v1.GET("/users", userHandler.List)

	log.Printf("api listening on %s", cfg.Address())
	if err := router.Run(cfg.Address()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
