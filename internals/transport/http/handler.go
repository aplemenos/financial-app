package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	log "github.com/sirupsen/logrus"
)

// Handler - stores pointer to our transaction service
type Handler struct {
	Router             *mux.Router
	TransactionService TransactionService
	Server             *http.Server
}

// Response object
type Response struct {
	Message string `json:"message"`
}

// NewHandler - returns a pointer to a Handler
func NewHandler(txnService TransactionService) *Handler {
	log.Info("setting up our handler")
	h := &Handler{
		TransactionService: txnService,
	}

	// Set up the routes
	router := h.mapRoutes()
	// Sets up our middleware functions
	router.Use(JSONMiddleware)
	// We want to timeout all requests that take longer than 15 seconds
	router.Use(TimeoutMiddleware)
	// We also want to log every incoming request with request id
	router.Use(LoggingMiddleware)

	// Get the timeouts from the enviroment variable
	rwTimeout, err := strconv.ParseInt(os.Getenv("RW_TIMEOUT"), 10, 0)
	if err != nil {
		panic(err)
	}
	rwt := time.Duration(rwTimeout) * time.Second

	idleTimeout, err := strconv.ParseInt(os.Getenv("IDLE_TIMEOUT"), 10, 0)
	if err != nil {
		panic(err)
	}
	idlet := time.Duration(idleTimeout) * time.Second

	h.Server = &http.Server{
		Addr: os.Getenv("SERVER_ADDR"),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: rwt,
		ReadTimeout:  rwt,
		IdleTimeout:  idlet,
		Handler:      router,
	}
	// return our wonderful handler
	return h
}

// mapRoutes - sets up all the routes for financial application
func (h *Handler) mapRoutes() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/alive", h.AliveCheck).Methods("GET")

	financialRoutes := router.PathPrefix("/api/v1").Subrouter()
	financialRoutes.HandleFunc("/transaction", h.Transfer).Methods("POST")
	financialRoutes.HandleFunc("/transaction/{id}", h.GetTransaction).Methods("GET")
	financialRoutes.HandleFunc("/transaction/{id}", h.DeleteTransaction).Methods("DELETE")

	financialRoutes.HandleFunc("/account", h.PostAccount).Methods("POST")
	financialRoutes.HandleFunc("/account/{id}", h.GetAccount).Methods("GET")
	financialRoutes.HandleFunc("/account/{id}", h.DeleteAccount).Methods("DELETE")

	return router
}

// Serve - gracefully serves our newly set up handler function
func (h *Handler) Serve() error {
	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			log.Error(err)
		}
	}()

	// Create a deadline to wait for
	serverTimeout, err := strconv.ParseInt(os.Getenv("SERVER_TIMEOUT"), 10, 0)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("the server timeout is ", serverTimeout)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	timeout := time.Duration(serverTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Shut downs gracefully the server
	if err := h.Server.Shutdown(ctx); err != nil {
		log.Error(err)
		return err
	}

	log.Info("shutting down gracefully")
	return nil
}
