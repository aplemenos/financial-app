package decoratedsvcs

import (
	"context"
	"financial-app/pkg/healthchecks"
	"time"

	log "go.uber.org/zap"
)

type loggingService struct {
	logger *log.SugaredLogger
	next   healthchecks.Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger *log.SugaredLogger, s healthchecks.Service) healthchecks.Service {
	return &loggingService{logger, s}
}

func (s *loggingService) Alive(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		s.logger.Infow(
			"alive",
			log.Duration("took", time.Since(begin)),
			log.Error(err),
		)
	}(time.Now())
	return s.next.Alive(ctx)
}
