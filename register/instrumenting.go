package register

import (
	"context"
	"financial-app/domain/account"
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

func (s *instrumentingService) LoadAccount(
	ctx context.Context, id account.AccountID,
) (account Account, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "load").Add(1)
		s.requestLatency.With("method", "load").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.LoadAccount(ctx, id)
}

func (s *instrumentingService) Register(
	ctx context.Context, acct account.Account,
) (account Account, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "register").Add(1)
		s.requestLatency.With("method", "register").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Register(ctx, acct)
}

func (s *instrumentingService) Accounts(ctx context.Context) []Account {
	defer func(begin time.Time) {
		s.requestCount.With("method", "accounts").Add(1)
		s.requestLatency.With("method", "accounts").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Accounts(ctx)
}

func (s *instrumentingService) Clean(
	ctx context.Context, id account.AccountID,
) (err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "clean").Add(1)
		s.requestLatency.With("method", "clean").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Clean(ctx, id)
}
