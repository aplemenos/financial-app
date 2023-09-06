package rest

import (
	"context"
	"financial-app/pkg/accounts"
	acctsvcs "financial-app/pkg/accounts/decoratedsvcs"
	"financial-app/pkg/healthchecks"
	healthsvcs "financial-app/pkg/healthchecks/decoratedsvcs"
	"financial-app/pkg/transactions"
	txnsvcs "financial-app/pkg/transactions/decoratedsvcs"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server holds the dependencies for a HTTP server.
type Server struct {
	AccountService     accounts.Service
	TransactionService transactions.Service
	HealthcheckService healthchecks.Service

	Logger *zap.SugaredLogger

	router *gin.Engine
}

// setupServices configures the financial app services
func setupServices(
	accountRepo accounts.AccountRepository,
	transactionRepo transactions.TransactionRepository,
	healthcheckRepo healthchecks.HealthcheckRepository,
	log *zap.SugaredLogger,
) (accounts.Service, transactions.Service, healthchecks.Service) {
	fieldKeys := []string{"method"}

	// Setup services
	var as accounts.Service
	as = accounts.NewService(accountRepo)
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

	var ts transactions.Service
	ts = transactions.NewService(accountRepo, transactionRepo)
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

	var hs healthchecks.Service
	hs = healthchecks.NewService(healthcheckRepo)
	hs = healthsvcs.NewLoggingService(log, hs)

	return as, ts, hs
}

// NewServer returns a new HTTP server.
func NewServer(
	accountRepo accounts.AccountRepository,
	transactionRepo transactions.TransactionRepository,
	healthcheckRepo healthchecks.HealthcheckRepository,
	logger *zap.SugaredLogger,
) *Server {
	as, ts, hs := setupServices(accountRepo, transactionRepo, healthcheckRepo, logger)
	s := &Server{
		AccountService:     as,
		TransactionService: ts,
		HealthcheckService: hs,
		Logger:             logger,
	}

	// Creates a router without any middleware by default
	r := gin.New()

	// Global middleware
	// Logger middleware will write the logs to gin.DefaultWriter
	// even if you set with GIN_MODE=release.
	// By default gin.DefaultWriter = os.Stdout
	r.Use(gin.Logger())
	// Recovery middleware recovers from any panics and writes a 500 if there was one.
	r.Use(gin.Recovery())
	// Custom middlewares
	r.Use(timeoutMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST, OPTIONS, GET, PUT, DELETE"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup routes
	servicesRoutes := r.Group("/api/v1/")

	// healthchecks
	hh := healthchecks.HealthcheckHandler{Service: s.HealthcheckService, Logger: s.Logger}
	hh.Router(&r.RouterGroup)
	// accounts
	ah := accounts.AccountHandler{Service: s.AccountService, Logger: s.Logger}
	ah.Router(servicesRoutes)
	// transactions
	th := transactions.TransactionHandler{Service: s.TransactionService, Logger: s.Logger}
	th.Router(servicesRoutes)
	// metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	s.router = r

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func timeoutResponse(c *gin.Context) {
	c.String(http.StatusRequestTimeout, "server timeout")
}

func timeoutMiddleware() gin.HandlerFunc {
	serverTimeout, _ := strconv.ParseInt(os.Getenv("SERVER_TIMEOUT"), 10, 0)
	return timeout.New(
		timeout.WithTimeout(time.Duration(serverTimeout)*time.Second),
		timeout.WithHandler(func(c *gin.Context) {
			c.Next()
		}),
		timeout.WithResponse(timeoutResponse),
	)
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	s.Logger.Info("shutdown server  ...")

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
