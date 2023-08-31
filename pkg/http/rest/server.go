package rest

import (
	"context"
	"financial-app/pkg/account"
	acctsvcs "financial-app/pkg/account/decoratedsvcs"
	"financial-app/pkg/healthcheck"
	healthsvcs "financial-app/pkg/healthcheck/decoratedsvcs"
	"financial-app/pkg/transaction"
	txnsvcs "financial-app/pkg/transaction/decoratedsvcs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server holds the dependencies for a HTTP server.
type Server struct {
	Account     account.Service
	Transaction transaction.Service
	Healthcheck healthcheck.Service

	Logger *zap.SugaredLogger

	router chi.Router
}

// setupServices configures the financial app services
func setupServices(
	accountRepo account.AccountRepository,
	transactionRepo transaction.TransactionRepository,
	healthcheckRepo healthcheck.HealthcheckRepository,
	log *zap.SugaredLogger,
) (account.Service, transaction.Service, healthcheck.Service) {
	fieldKeys := []string{"method"}

	// Setup services
	var as account.Service
	as = account.NewService(accountRepo)
	as = acctsvcs.NewLoggingService(log, as)
	as = acctsvcs.NewInstrumentingService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "account_service",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "account_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		as)

	var ts transaction.Service
	ts = transaction.NewService(accountRepo, transactionRepo)
	ts = txnsvcs.NewLoggingService(log, ts)
	ts = txnsvcs.NewInstrumentingService(
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "transaction_service",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, fieldKeys),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "api",
			Subsystem: "transaction_service",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, fieldKeys),
		ts)

	var hs healthcheck.Service
	hs = healthcheck.NewService(healthcheckRepo)
	hs = healthsvcs.NewLoggingService(log, hs)

	return as, ts, hs
}

// NewServer returns a new HTTP server.
func NewServer(
	accountRepo account.AccountRepository,
	transactionRepo transaction.TransactionRepository,
	healthcheckRepo healthcheck.HealthcheckRepository,
	logger *zap.SugaredLogger,
) *Server {
	as, ts, hs := setupServices(accountRepo, transactionRepo, healthcheckRepo, logger)
	s := &Server{
		Account:     as,
		Transaction: ts,
		Healthcheck: hs,
		Logger:      logger,
	}

	r := chi.NewRouter()

	r.Use(s.accessControl)
	r.Use(s.jsonMiddleware)
	r.Use(s.timeoutMiddleware)
	r.Use(s.recovery)

	r.Route("/api/v1", func(r chi.Router) {
		ah := account.AccountHandler{Service: s.Account, Logger: s.Logger}
		th := transaction.TransactionHandler{Service: s.Transaction, Logger: s.Logger}
		r.Mount("/accounts", ah.Router())
		r.Mount("/transactions", th.Router())
	})

	hh := healthcheck.HealthcheckHandler{Service: s.Healthcheck, Logger: s.Logger}
	r.Mount("/alive", hh.Router())

	r.Method("GET", "/metrics", promhttp.Handler())

	s.router = r

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) jsonMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		h.ServeHTTP(w, r)
	})
}

func (s *Server) timeoutMiddleware(h http.Handler) http.Handler {
	timeout := os.Getenv("SERVER_TIMEOUT")
	serverTimeout, _ := strconv.ParseInt(timeout, 10, 0)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Duration(serverTimeout)*time.Second)
		defer cancel()
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// recovery is a wrapper which will try to recover from any panic error and report it
func (s *Server) recovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			err := recover()
			if err != nil {
				// Log the error in order to know the failure's reason
				s.Logger.Error(err)

				http.Error(w, "There was an internal server error", http.StatusInternalServerError)
			}
		}()

		h.ServeHTTP(w, r)
	})
}

func (s *Server) accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

// Serve gracefully serves our newly set up handler function
func (s *Server) Serve(server *http.Server, stimeout string) error {
	go func() {
		if err := server.ListenAndServe(); err != nil {
			s.Logger.Error(err)
		}
	}()

	// Create a deadline to wait for
	serverTimeout, err := strconv.ParseInt(stimeout, 10, 0)
	if err != nil {
		s.Logger.Error(err)
		return err
	}
	s.Logger.Debug("the server timeout is ", serverTimeout)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	timeout := time.Duration(serverTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Shut downs gracefully the server
	if err := server.Shutdown(ctx); err != nil {
		s.Logger.Error(err)
		return err
	}

	s.Logger.Info("shutting down gracefully")
	return nil
}
