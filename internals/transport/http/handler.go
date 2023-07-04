package http

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
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

	h.Server = &http.Server{
		Addr: "0.0.0.0:8080",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      h.Router,
	}
	// return our wonderful handler
	return h
}

// mapRoutes - sets up all the routes for financial application
func (h *Handler) mapRoutes() {
	h.Router.HandleFunc("/alive", h.AliveCheck).Methods("GET")
	h.Router.HandleFunc("/ready", h.ReadyCheck).Methods("GET")
	h.Router.HandleFunc("/api/v1/transaction", h.PostTransaction).Methods("POST")
	h.Router.HandleFunc("/api/v1/transaction/{id}", h.GetTransaction).Methods("GET")
	h.Router.HandleFunc("/api/v1/transaction/{id}", h.DeleteTransaction).Methods("DELETE")

	h.Router.HandleFunc("/api/v1/account", h.PostAccount).Methods("POST")
	h.Router.HandleFunc("/api/v1/account/{id}", h.UpdateAccount).Methods("PUT")
	h.Router.HandleFunc("/api/v1/account/{id}", h.GetAccount).Methods("GET")
	h.Router.HandleFunc("/api/v1/account/{id}", h.DeleteAccount).Methods("DELETE")

}

func (h *Handler) AliveCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Response{Message: "I am Alive!"}); err != nil {
		panic(err)
	}
}

// Serve - gracefully serves our newly set up handler function
func (h *Handler) Serve() error {
	go func() {
		if err := h.Server.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := h.Server.Shutdown(ctx); err != nil {
		log.Println(err)
		return err
	}

	log.Println("shutting down gracefully")
	return nil
}
