package main

import (
	"financial-app/pkg/http/rest"
	"financial-app/pkg/postgres"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const (
	defaultServerAddr    = "0.0.0.0:8080"
	defaultRWTimeout     = "15"
	defaultIdleTimeout   = "15"
	defaultServerTimeout = "15"
	defaultDBName        = "postgres"
	defaultDBPassword    = "postgres"
	defaultDBHost        = "db"
	defaultDBTable       = "postgres"
	defaultDBPort        = "5432"
	defaultSSLMode       = "disable"
)

// run sets up our application
func run() error {
	// Build a production logger
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	log := logger.Sugar()

	log.Info("setting up financial app")

	// Setup the postgres DB
	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		envString("DB_HOST", defaultDBHost),
		envString("DB_PORT", defaultDBPort),
		envString("DB_USERNAME", defaultDBName),
		envString("DB_TABLE", defaultDBTable),
		envString("DB_PASSWORD", defaultDBPassword),
		envString("SSL_MODE", defaultSSLMode),
	)

	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		log.Error("failed to connect to database")
		return err
	}

	// Setup the repositories
	accountRepo := postgres.NewAccountRepository(db.DB, log)
	transactionRepo := postgres.NewTransactionRepository(db.DB, log)
	healthRepo := postgres.NewHealthcheckRepository(db.DB, log)

	// Setup the server
	srv := rest.NewServer(accountRepo, transactionRepo, healthRepo, log)

	// Run the server
	serverConfig, err := loadServerSettings(srv)
	if err != nil {
		log.Error(err)
		return err
	}

	if err := srv.Serve(serverConfig,
		envString("SERVER_TIMEOUT", defaultServerTimeout)); err != nil {
		log.Error("failed to gracefully serve financial app")
		return err
	}

	return nil
}

func loadServerSettings(srv *rest.Server) (*http.Server, error) {
	// Get the timeouts from the enviroment variable
	rwTimeout, err := strconv.ParseInt(envString("RW_TIMEOUT", defaultRWTimeout), 10, 0)
	if err != nil {
		return nil, err
	}
	rwt := time.Duration(rwTimeout) * time.Second

	idleTimeout, err := strconv.ParseInt(envString("IDLE_TIMEOUT", defaultIdleTimeout), 10, 0)
	if err != nil {
		return nil, err
	}
	idlet := time.Duration(idleTimeout) * time.Second

	serverConfig := &http.Server{
		Addr: envString("SERVER_ADDR", defaultServerAddr),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: rwt,
		ReadTimeout:  rwt,
		IdleTimeout:  idlet,
		Handler:      srv,
	}

	return serverConfig, nil
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

func main() {
	if err := run(); err != nil {
		zap.S().Error(err)
		zap.S().Panic("Error starting up financial app")
	}
}
