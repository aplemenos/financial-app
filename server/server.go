package server

import (
	"context"
	"encoding/json"
	"financial-app/register"
	"financial-app/transfer"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// response object
type response struct {
	Message string `json:"message"`
}

// Server holds the dependencies for a HTTP server.
type Server struct {
	Account  register.Service
	Transfer transfer.Service

	Logger *zap.SugaredLogger

	router chi.Router
}

// New returns a new HTTP server.
func New(as register.Service, ts transfer.Service, logger *zap.SugaredLogger) *Server {
	s := &Server{
		Account:  as,
		Transfer: ts,
		Logger:   logger,
	}

	r := chi.NewRouter()

	r.Use(s.accessControl)
	r.Use(s.jsonMiddleware)
	r.Use(s.timeoutMiddleware)
	r.Use(s.recovery)

	r.Route("/api/v1", func(r chi.Router) {
		ah := registerHandler{s.Account, s.Logger}
		th := transferHandler{s.Transfer, s.Logger}
		r.Mount("/accounts", ah.router())
		r.Mount("/transactions", th.router())
	})

	r.Get("/alive", s.aliveCheck)

	r.Method("GET", "/metrics", promhttp.Handler())

	s.router = r

	return s
}

func (s *Server) aliveCheck(w http.ResponseWriter, r *http.Request) {
	if err := s.Transfer.Alive(r.Context()); err != nil {
		s.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response{Message: "I am Alive!"}); err != nil {
		s.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
				s.Logger.Error(err)

				w.WriteHeader(http.StatusInternalServerError)

				json.NewEncoder(w).Encode(map[string]string{
					"status": "error",
					"desc":   "There was an internal server error",
				})
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
