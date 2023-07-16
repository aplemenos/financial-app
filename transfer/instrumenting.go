package transfer

import (
	"context"
	"financial-app/domain/transaction"
	"time"

	"github.com/go-kit/kit/metrics"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           Service
}

// NewInstrumentingService returns an instance of an instrumenting Service.
func NewInstrumentingService(
	counter metrics.Counter, latency metrics.Histogram, s Service,
) Service {
	return &instrumentingService{
		requestCount:   counter,
		requestLatency: latency,
		next:           s,
	}
}

func (s *instrumentingService) LoadTransaction(
	ctx context.Context, id transaction.TransactionID,
) (transaction Transaction, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "load").Add(1)
		s.requestLatency.With("method", "load").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.LoadTransaction(ctx, id)
}

func (s *instrumentingService) Transfer(
	ctx context.Context, txn transaction.Transaction,
) (transaction Transaction, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "transfer").Add(1)
		s.requestLatency.With("method", "transfer").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Transfer(ctx, txn)
}

func (s *instrumentingService) Transactions(ctx context.Context) []Transaction {
	defer func(begin time.Time) {
		s.requestCount.With("method", "transactions").Add(1)
		s.requestLatency.With("method", "transactions").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Transactions(ctx)
}

func (s *instrumentingService) Clean(
	ctx context.Context, id transaction.TransactionID,
) (err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "clean").Add(1)
		s.requestLatency.With("method", "clean").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Clean(ctx, id)
}

func (s *instrumentingService) Alive(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "alive").Add(1)
		s.requestLatency.With("method", "alive").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Alive(ctx)
}
