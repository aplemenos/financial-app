package main

import (
	"financial-app/domain/account"
	"financial-app/domain/transaction"
	"financial-app/multiplelock"
	"financial-app/postgres"
	"financial-app/register"
	"financial-app/server"
	"financial-app/transfer"
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

// Run - sets up our application
func Run() error {
	// Build a production logger
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	log := logger.Sugar()

	log.Info("setting up financial app")

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

	err = postgres.MigrateDB(db.DB)
	if err != nil {
		log.Error("failed to setup database")
		return err
	}

	// Setup repositories
	var (
		accounts     account.AccountRepository
		transactions transaction.TransactionRepository
	)
	accounts = postgres.NewAccountRepository(db.DB, log)
	transactions = postgres.NewTransactionRepository(db.DB, log)

	// Setup services
	var rs register.Service
	rs = register.NewService(accounts)
	rs = register.NewLoggingService(log, rs)

	var ts transfer.Service
	mlock := multiplelock.NewMultipleLock()
	ts = transfer.NewService(accounts, transactions, mlock)
	ts = transfer.NewLoggingService(log, ts)

	srv := server.New(rs, ts, log)

	// Get the timeouts from the enviroment variable
	rwTimeout, err := strconv.ParseInt(envString("RW_TIMEOUT", defaultRWTimeout), 10, 0)
	if err != nil {
		log.Error("failed to parse RW_TIMEOUT")
		return err
	}
	rwt := time.Duration(rwTimeout) * time.Second

	idleTimeout, err := strconv.ParseInt(envString("IDLE_TIMEOUT", defaultIdleTimeout), 10, 0)
	if err != nil {
		log.Error("failed to parse IDLE_TIMEOUT")
		return err
	}
	idlet := time.Duration(idleTimeout) * time.Second

	server := &http.Server{
		Addr: envString("SERVER_ADDR", defaultServerAddr),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: rwt,
		ReadTimeout:  rwt,
		IdleTimeout:  idlet,
		Handler:      srv,
	}

	if err := srv.Serve(server, envString("SERVER_TIMEOUT", defaultServerTimeout)); err != nil {
		log.Error("failed to gracefully serve financial app")
		return err
	}

	return nil
}

func main() {
	if err := Run(); err != nil {
		zap.S().Error(err)
		zap.S().Panic("Error starting up financial app")
	}
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}
