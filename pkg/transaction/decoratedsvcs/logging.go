package decoratedsvcs

import (
	"context"
	"financial-app/pkg/transaction"
	"time"

	log "go.uber.org/zap"
)

type loggingService struct {
	logger *log.SugaredLogger
	next   transaction.Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger *log.SugaredLogger, s transaction.Service) transaction.Service {
	return &loggingService{logger, s}
}

func (s *loggingService) Load(
	ctx context.Context, id string,
) (transaction transaction.Transaction, err error) {
	defer func(begin time.Time) {
		s.logger.Infow(
			"load",
			log.String("transaction_id", string(id)),
			log.Duration("took", time.Since(begin)),
			log.Error(err),
		)
	}(time.Now())
	return s.next.Load(ctx, id)
}

func (s *loggingService) Transfer(
	ctx context.Context, txn transaction.Transaction,
) (transaction transaction.Transaction, err error) {
	defer func(begin time.Time) {
		s.logger.Infow(
			"tranfer",
			log.String("transaction_id", string(txn.ID)),
			log.String("source_account_id", string(txn.SourceAccountID)),
			log.String("target_account_id", string(txn.TargetAccountID)),
			log.Float64("amount", txn.Amount),
			log.String("currency", string(txn.Currency)),
			log.Duration("took", time.Since(begin)),
			log.Error(err),
		)
	}(time.Now())
	return s.next.Transfer(ctx, txn)
}

func (s *loggingService) LoadAll(ctx context.Context) []transaction.Transaction {
	defer func(begin time.Time) {
		s.logger.Infow(
			"loadall",
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
			log.String("transaction_id", string(id)),
			log.Duration("took", time.Since(begin)),
			log.Error(err),
		)
	}(time.Now())
	return s.next.Clean(ctx, id)
}
