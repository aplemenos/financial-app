package decoratedsvcs

import (
	"context"
	"financial-app/pkg/accounts"
	"time"

	log "go.uber.org/zap"
)

type loggingService struct {
	logger *log.SugaredLogger
	next   accounts.Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger *log.SugaredLogger, s accounts.Service) accounts.Service {
	return &loggingService{logger, s}
}

func (s *loggingService) Load(
	ctx context.Context, id string,
) (account accounts.Account, err error) {
	defer func(begin time.Time) {
		s.logger.Infow(
			"load",
			log.String("account_id", string(id)),
			log.Duration("took", time.Since(begin)),
			log.Error(err),
		)
	}(time.Now())
	return s.next.Load(ctx, id)
}

func (s *loggingService) Register(
	ctx context.Context, acct accounts.Account,
) (account accounts.Account, err error) {
	defer func(begin time.Time) {
		s.logger.Infow(
			"account store",
			log.String("account_id", string(acct.ID)),
			log.Float64("balance", acct.Balance),
			log.String("currency", string(acct.Currency)),
			log.String("created_at", acct.CreatedAt.String()),
			log.Duration("took", time.Since(begin)),
			log.Error(err),
		)
	}(time.Now())
	return s.next.Register(ctx, acct)
}

func (s *loggingService) LoadAll(ctx context.Context) []accounts.Account {
	defer func(begin time.Time) {
		s.logger.Infow(
			"accounts",
			log.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return s.next.LoadAll(ctx)
}

func (s *loggingService) Clean(
	ctx context.Context, id string,
) (err error) {
	defer func(begin time.Time) {
		s.logger.Infow(
			"clean",
			log.String("account_id", string(id)),
			log.Duration("took", time.Since(begin)),
			log.Error(err),
		)
	}(time.Now())
	return s.next.Clean(ctx, id)
}
