package scheduler

import "log/slog"

type Scheduler struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Scheduler {
	return &Scheduler{logger: logger}
}

func (s *Scheduler) Start() {
	s.logger.Info("worker scheduler started")
}
