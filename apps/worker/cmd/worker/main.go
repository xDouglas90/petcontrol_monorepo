package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/config"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/mail"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/processor"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/queue"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/scheduler"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/whatsapp"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load worker config", "error", err.Error())
		os.Exit(1)
	}

	if err := pingRedis(cfg.RedisAddr); err != nil {
		logger.Error("redis connection failed", "redis_addr", cfg.RedisAddr, "error", err.Error())
		os.Exit(1)
	}

	wa := whatsapp.NewClient(logger)
	emailSender := mail.NewSMTPSender(logger, cfg)
	notifProcessor := processor.NewNotificationsProcessor(logger, wa)
	scheduleProcessor := processor.NewScheduleConfirmationProcessor(logger, wa)
	peopleAccessProcessor := processor.NewPersonAccessCredentialsProcessor(logger, emailSender)
	scheduler.New(logger).Start()
	webhookServer := &http.Server{
		Addr:    cfg.WebhookAddr,
		Handler: whatsapp.NewWebhookHandler(cfg.WhatsAppVerifyToken, logger),
	}

	go func() {
		logger.Info("worker webhook started", "addr", cfg.WebhookAddr)
		if err := webhookServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("worker webhook stopped unexpectedly", "error", err.Error())
			os.Exit(1)
		}
	}()

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeNotificationDummy, notifProcessor.HandleDummyNotification)
	mux.HandleFunc(queue.TypeScheduleConfirmed, scheduleProcessor.HandleScheduleConfirmation)
	mux.HandleFunc(queue.TypePersonAccessCredentials, peopleAccessProcessor.HandlePersonAccessCredentials)

	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			Concurrency: cfg.Concurrency,
			Queues: map[string]int{
				cfg.WorkerQueue: 1,
			},
		},
	)

	go func() {
		logger.Info("worker started", "redis_addr", cfg.RedisAddr, "queue", cfg.WorkerQueue)
		if err := server.Run(mux); err != nil {
			logger.Error("worker stopped unexpectedly", "error", err.Error())
			os.Exit(1)
		}
	}()

	waitForShutdownSignal(logger)
	if err := webhookServer.Shutdown(context.Background()); err != nil {
		logger.Error("worker webhook shutdown failed", "error", err.Error())
	}
	server.Shutdown()
}

func pingRedis(addr string) error {
	client := redis.NewClient(&redis.Options{Addr: addr})
	defer client.Close()
	return client.Ping(context.Background()).Err()
}

func waitForShutdownSignal(logger *slog.Logger) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
	logger.Info("worker shutdown requested")
}
