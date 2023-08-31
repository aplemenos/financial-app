package healthcheck

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// response object
type response struct {
	Message string `json:"message"`
}

type HealthcheckHandler struct {
	Service Service

	Logger *zap.SugaredLogger
}

// router sets up all the routes for healthcheck service
func (h *HealthcheckHandler) Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.aliveCheck)

	return r
}

func (h *HealthcheckHandler) aliveCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.Service.Alive(r.Context()); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response{Message: "I am Alive!"}); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
