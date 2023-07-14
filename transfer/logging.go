package transfer

import (
	"context"
	"financial-app/domain/transaction"
	"time"

	log "go.uber.org/zap"
)

type loggingService struct {
	logger *log.SugaredLogger
	next   Service
}

// NewLoggingService returns a new instance of a logging Service.
func NewLoggingService(logger *log.SugaredLogger, s Service) Service {
	return &loggingService{logger, s}
}

func (s *loggingService) LoadTransaction(
	ctx context.Context, id transaction.TransactionID,
) (transaction Transaction, err error) {
	defer func(begin time.Time) {
		s.logger.Infow(
			"load",
			log.String("transaction_id", string(id)),
			log.Duration("took", time.Since(begin)),
			log.Error(err),
		)
	}(time.Now())
	return s.next.LoadTransaction(ctx, id)
}

func (s *loggingService) Transfer(
	ctx context.Context, txn transaction.Transaction,
) (transaction Transaction, err error) {
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

func (s *loggingService) Transactions(ctx context.Context) []Transaction {
	defer func(begin time.Time) {
		s.logger.Infow(
			"transactions",
			log.Duration("took", time.Since(begin)),
		)
	}(time.Now())
	return s.next.Transactions(ctx)
}

func (s *loggingService) Clean(
	ctx context.Context, id transaction.TransactionID,
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
