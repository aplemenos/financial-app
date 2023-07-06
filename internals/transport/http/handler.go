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

	h.Router = mux.NewRouter()
	// Sets up our middleware functions
	h.Router.Use(JSONMiddleware)
	// We want to timeout all requests that take longer than 15 seconds
	h.Router.Use(TimeoutMiddleware)
	// set up the routes
	h.mapRoutes()

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
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: rwt,
		ReadTimeout:  rwt,
		IdleTimeout:  idlet,
		Handler:      h.Router,
	}
	// return our wonderful handler
	return h
}

// mapRoutes - sets up all the routes for financial application
func (h *Handler) mapRoutes() {
	h.Router.HandleFunc("/alive", h.AliveCheck).Methods("GET")
	h.Router.HandleFunc("/api/v1/transaction", h.Transfer).Methods("POST")
	h.Router.HandleFunc("/api/v1/transaction/{id}", h.GetTransaction).Methods("GET")
	h.Router.HandleFunc("/api/v1/transaction/{id}", h.DeleteTransaction).Methods("DELETE")

	h.Router.HandleFunc("/api/v1/account", h.PostAccount).Methods("POST")
	h.Router.HandleFunc("/api/v1/account/{id}", h.GetAccount).Methods("GET")
	h.Router.HandleFunc("/api/v1/account/{id}", h.DeleteAccount).Methods("DELETE")

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

	if err := h.Server.Shutdown(ctx); err != nil {
		log.Error(err)
		return err
	}

	log.Info("shutting down gracefully")
	return nil
}
