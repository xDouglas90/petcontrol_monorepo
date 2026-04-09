package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/xdouglas90/petcontrol_monorepo/worker/internal/config"
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
	notifProcessor := processor.NewNotificationsProcessor(logger, wa)
	scheduler.New(logger).Start()

	mux := asynq.NewServeMux()
	mux.HandleFunc(queue.TypeNotificationDummy, notifProcessor.HandleDummyNotification)

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
