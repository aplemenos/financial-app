package decoratedsvcs

import (
	"context"
	"financial-app/pkg/accounts"
	"time"

	"github.com/go-kit/kit/metrics"
)

type instrumentingService struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	next           accounts.Service
}

// NewInstrumentingService returns an instance of an instrumenting Service.
func NewInstrumentingService(
	counter metrics.Counter, latency metrics.Histogram, s accounts.Service,
) accounts.Service {
	return &instrumentingService{
		requestCount:   counter,
		requestLatency: latency,
		next:           s,
	}
}

func (s *instrumentingService) Load(
	ctx context.Context, id string,
) (account accounts.Account, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "load").Add(1)
		s.requestLatency.With("method", "load").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Load(ctx, id)
}

func (s *instrumentingService) Register(
	ctx context.Context, acct accounts.Account,
) (account accounts.Account, err error) {
	defer func(begin time.Time) {
		s.requestCount.With("method", "register").Add(1)
		s.requestLatency.With("method", "register").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return s.next.Register(ctx, acct)
}

func (s *instrumentingService) LoadAll(ctx context.Context) []accounts.Account {
	defer func(begin time.Time) {
		s.requestCount.With("method", "accounts").Add(1)
		s.requestLatency.With("method", "accounts").Observe(time.Since(begin).Seconds())
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
