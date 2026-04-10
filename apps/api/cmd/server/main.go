package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xdouglas90/petcontrol_monorepo/internal/config"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db"
	"github.com/xdouglas90/petcontrol_monorepo/internal/db/sqlc"
	"github.com/xdouglas90/petcontrol_monorepo/internal/handler"
	"github.com/xdouglas90/petcontrol_monorepo/internal/middleware"
	"github.com/xdouglas90/petcontrol_monorepo/internal/queue"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
)

// @title PetControl API
// @version 1.0
// @description API HTTP do PetControl para autenticação, módulos de tenant e agendamentos.
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Use o formato: Bearer <token>

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
	companyService := service.NewCompanyService(queries)
	planService := service.NewPlanService(queries)
	moduleService := service.NewModuleService(queries)
	companyUserService := service.NewCompanyUserService(queries)
	scheduleService := service.NewScheduleService(queries)
	authService := service.NewAuthService(queries, cfg.JWTSecret, cfg.JWTTTL)
	workerPublisher := queue.NewAsynqPublisher(cfg.RedisAddr, cfg.WorkerQueue)
	defer func() {
		if err := workerPublisher.Close(); err != nil {
			log.Printf("close worker publisher: %v", err)
		}
	}()
	companyHandler := handler.NewCompanyHandler(companyService)
	planHandler := handler.NewPlanHandler(planService)
	moduleHandler := handler.NewModuleHandler(moduleService)
	companyUserHandler := handler.NewCompanyUserHandler(companyUserService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService, workerPublisher)
	authHandler := handler.NewAuthHandler(authService)
	workerHandler := handler.NewWorkerHandler(workerPublisher)
	healthHandler := handler.NewHealthHandler(pool)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	router := gin.New()
	router.Use(
		middleware.RequestContext(),
		middleware.CORS(cfg.CORSAllowedOrigins),
		middleware.RequestLogger(logger),
		middleware.Recovery(logger),
	)

	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)
	registerSwaggerRoute(router)

	v1 := router.Group("/api/v1")
	v1.POST("/auth/login", authHandler.Login)

	protected := v1.Group("/")
	protected.Use(middleware.Auth(cfg.JWTSecret), middleware.Tenant(), middleware.Audit(queries, logger))
	protected.GET("/companies/current", companyHandler.Current)
	protected.GET("/plans/current", planHandler.Current)
	protected.GET("/modules/active", moduleHandler.Active)
	protected.GET("/company-users", companyUserHandler.List)
	protected.POST("/company-users", middleware.RequireCompanyOwner(queries), companyUserHandler.Create)
	protected.DELETE("/company-users/:user_id", middleware.RequireCompanyOwner(queries), companyUserHandler.Deactivate)
	protected.POST("/worker/notifications/dummy", workerHandler.EnqueueDummyNotification)
	protected.GET("/modules/:code/access", middleware.RequireModule(queries, ""), func(c *gin.Context) {
		middleware.JSONData(c, 200, gin.H{"allowed": true, "module": c.Param("code")})
	})

	schedules := protected.Group("/schedules")
	schedules.Use(middleware.RequireModule(queries, "SCH"))
	schedules.GET("", scheduleHandler.List)
	schedules.POST("", scheduleHandler.Create)
	schedules.GET("/:id", scheduleHandler.GetByID)
	schedules.GET("/:id/history", scheduleHandler.History)
	schedules.PUT("/:id", scheduleHandler.Update)
	schedules.DELETE("/:id", scheduleHandler.Delete)

	log.Printf("api listening on %s", cfg.Address())
	if err := router.Run(cfg.Address()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
