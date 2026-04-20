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
	"github.com/xdouglas90/petcontrol_monorepo/internal/realtime"
	"github.com/xdouglas90/petcontrol_monorepo/internal/service"
	"github.com/xdouglas90/petcontrol_monorepo/internal/storage/gcs"
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
	gcsClient, err := gcs.NewClient(context.Background(), cfg.Uploads)
	if err != nil {
		log.Fatalf("gcs client initialization failed: %v", err)
	}
	if gcsClient != nil {
		defer func() {
			if err := gcsClient.Close(); err != nil {
				log.Printf("close gcs client: %v", err)
			}
		}()
	}

	companyService := service.NewCompanyService(queries)
	companySystemConfigService := service.NewCompanySystemConfigService(queries)
	planService := service.NewPlanService(queries)
	moduleService := service.NewModuleService(queries)
	userService := service.NewUserService(queries)
	companyUserService := service.NewCompanyUserService(queries)
	adminSystemChatService := service.NewAdminSystemChatService(queries)
	internalChatHub := realtime.NewInternalChatHub()
	clientService := service.NewClientService(pool, queries)
	petService := service.NewPetService(queries)
	serviceService := service.NewServiceService(pool, queries)
	scheduleService := service.NewScheduleService(pool, queries)
	authService := service.NewAuthService(queries, cfg.JWTSecret, cfg.JWTTTL)
	uploadStorage := gcs.NewService(cfg.Uploads, gcsClient, gcsClient)
	uploadService := service.NewUploadService(uploadStorage)
	workerPublisher := queue.NewAsynqPublisher(cfg.RedisAddr, cfg.WorkerQueue)
	defer func() {
		if err := workerPublisher.Close(); err != nil {
			log.Printf("close worker publisher: %v", err)
		}
	}()
	companyHandler := handler.NewCompanyHandler(companyService, uploadService)
	companySystemConfigHandler := handler.NewCompanySystemConfigHandler(companySystemConfigService)
	planHandler := handler.NewPlanHandler(planService)
	moduleHandler := handler.NewModuleHandler(moduleService)
	userHandler := handler.NewUserHandler(userService)
	companyUserHandler := handler.NewCompanyUserHandler(companyUserService)
	adminSystemChatHandler := handler.NewAdminSystemChatHandler(adminSystemChatService, internalChatHub, cfg.CORSAllowedOrigins)
	clientHandler := handler.NewClientHandler(clientService, uploadService)
	petHandler := handler.NewPetHandler(petService, uploadService)
	serviceHandler := handler.NewServiceHandler(serviceService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService, workerPublisher)
	authHandler := handler.NewAuthHandler(authService)
	uploadHandler := handler.NewUploadHandler(uploadService)
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
	protected.PATCH("/companies/current", companyHandler.Update)
	protected.GET("/company-system-configs/current", companySystemConfigHandler.Current)
	protected.PATCH("/company-system-configs/current", companySystemConfigHandler.Update)
	protected.GET("/users/me", userHandler.Current)
	protected.GET("/plans/current", planHandler.Current)
	protected.GET("/modules/active", moduleHandler.Active)
	protected.GET("/company-users", companyUserHandler.List)
	protected.GET("/chat/system/:user_id/messages", adminSystemChatHandler.ListMessages)
	protected.POST("/chat/system/:user_id/messages", adminSystemChatHandler.CreateMessage)
	protected.POST("/company-users", middleware.RequireCompanyOwner(queries), companyUserHandler.Create)
	protected.DELETE("/company-users/:user_id", middleware.RequireCompanyOwner(queries), companyUserHandler.Deactivate)
	protected.POST("/worker/notifications/dummy", workerHandler.EnqueueDummyNotification)
	protected.GET("/modules/:code/access", middleware.RequireModule(queries, ""), moduleHandler.Access)
	protected.POST("/uploads/intent", uploadHandler.CreateIntent)
	protected.POST("/uploads/complete", uploadHandler.Complete)

	clients := protected.Group("/clients")
	clients.Use(middleware.RequireModule(queries, "CRM"))
	clients.GET("", clientHandler.List)
	clients.POST("", clientHandler.Create)
	clients.GET("/:id", clientHandler.GetByID)
	clients.PUT("/:id", clientHandler.Update)
	clients.DELETE("/:id", clientHandler.Delete)

	pets := protected.Group("/pets")
	pets.Use(middleware.RequireModule(queries, "CRM"))
	pets.GET("", petHandler.List)
	pets.POST("", petHandler.Create)
	pets.GET("/:id", petHandler.GetByID)
	pets.PUT("/:id", petHandler.Update)
	pets.DELETE("/:id", petHandler.Delete)

	services := protected.Group("/services")
	services.Use(middleware.RequireModule(queries, "SCH"))
	services.GET("", serviceHandler.List)
	services.POST("", serviceHandler.Create)
	services.GET("/:id", serviceHandler.GetByID)
	services.PUT("/:id", serviceHandler.Update)
	services.DELETE("/:id", serviceHandler.Delete)

	schedules := protected.Group("/schedules")
	schedules.Use(middleware.RequireModule(queries, "SCH"))
	schedules.GET("", scheduleHandler.List)
	schedules.POST("", scheduleHandler.Create)
	schedules.GET("/:id", scheduleHandler.GetByID)
	schedules.GET("/:id/history", scheduleHandler.History)
	schedules.PUT("/:id", scheduleHandler.Update)
	schedules.DELETE("/:id", scheduleHandler.Delete)

	protectedRealtime := v1.Group("/")
	protectedRealtime.Use(middleware.Auth(cfg.JWTSecret), middleware.Tenant())
	protectedRealtime.GET("/chat/system/:user_id/ws", adminSystemChatHandler.Connect)

	log.Printf("api listening on %s", cfg.Address())
	if err := router.Run(cfg.Address()); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
