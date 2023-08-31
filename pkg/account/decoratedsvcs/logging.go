package decoratedsvcs

import (
	"context"
	"financial-app/pkg/account"
	"time"

	log "go.uber.org/zap"
)

type loggingService struct {
	logger *log.SugaredLogger
	next   account.Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger *log.SugaredLogger, s account.Service) account.Service {
	return &loggingService{logger, s}
}

func (s *loggingService) LoadAccount(
	ctx context.Context, id string,
) (account account.Account, err error) {
	defer func(begin time.Time) {
		s.logger.Infow(
			"load",
			log.String("account_id", string(id)),
			log.Duration("took", time.Since(begin)),
			log.Error(err),
		)
	}(time.Now())
	return s.next.LoadAccount(ctx, id)
}

func (s *loggingService) Register(
	ctx context.Context, acct account.Account,
) (account account.Account, err error) {
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

func (s *loggingService) Accounts(ctx context.Context) []account.Account {
	defer func(begin time.Time) {
		s.logger.Infow(
			"accounts",
			log.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return s.next.Accounts(ctx)
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
