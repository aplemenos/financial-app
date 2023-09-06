package decoratedsvcs

import (
	"context"
	"financial-app/pkg/transactions"
	"time"

	"github.com/go-kit/kit/metrics"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           transactions.Service
}

// NewInstrumentingService returns an instance of an instrumenting Service.
func NewInstrumentingService(
	counter metrics.Counter, latency metrics.Histogram, s transactions.Service,
) transactions.Service {
	return &instrumentingService{
		requestCount:   counter,
		requestLatency: latency,
		next:           s,
	}
}

func (s *instrumentingService) Load(
	ctx context.Context, id string,
) (transaction transactions.Transaction, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "load").Add(1)
		s.requestLatency.With("method", "load").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Load(ctx, id)
}

func (s *instrumentingService) Transfer(
	ctx context.Context, txn transactions.Transaction,
) (transaction transactions.Transaction, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "transfer").Add(1)
		s.requestLatency.With("method", "transfer").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Transfer(ctx, txn)
}

func (s *instrumentingService) LoadAll(ctx context.Context) []transactions.Transaction {
	defer func(begin time.Time) {
		s.requestCount.With("method", "loadall").Add(1)
		s.requestLatency.With("method", "loadall").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.LoadAll(ctx)
}

func (s *instrumentingService) Clean(
	ctx context.Context, id string,
) (err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "clean").Add(1)
		s.requestLatency.With("method", "clean").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Clean(ctx, id)
}
