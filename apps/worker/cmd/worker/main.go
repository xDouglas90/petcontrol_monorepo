package main

import (
	"context"
	"log"
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
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load worker config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := pingRedis(cfg.RedisAddr); err != nil {
		log.Fatalf("redis connection failed: %v", err)
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
			log.Fatalf("worker stopped: %v", err)
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
