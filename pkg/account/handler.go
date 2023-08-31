package account

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var (
	accountIDRequired    = "account id required"
	currencyNotSupported = "currency is not supported"
)

type AccountHandler struct {
	Service Service

	Logger *zap.SugaredLogger
}

// Router sets up all the routes for account service
func (h *AccountHandler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.registerAccount)
	r.Get("/", h.accounts)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.loadAccount)
		r.Delete("/", h.clean)
	})

	return r
}

// loadAccount retrieves an account by ID
func (h *AccountHandler) loadAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.Logger.Error("no account id found")
		http.Error(w, accountIDRequired, http.StatusBadRequest)
		return
	}

	acct, err := h.Service.LoadAccount(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(acct); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// accounts retrieves all the registered accounts
func (h *AccountHandler) accounts(w http.ResponseWriter, r *http.Request) {
	accounts := h.Service.Accounts(r.Context())

	if err := json.NewEncoder(w).Encode(accounts); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// validCurrency validates if the given currency is supported
func validCurrency(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return IsSupportedCurrency(currency)
	}
	return false
}

// storeRequest
type storeRequest struct {
	Balance  float64 `json:"balance" validate:"required"`
	Currency string  `json:"currency" validate:"currency"`
}

func accountRequestFromAccountDomain(p storeRequest) Account {
	return Account{
		ID:       nextAccountID(), // Generate a new uuid
		Balance:  p.Balance,
		Currency: p.Currency,
	}
}

// registerAccount registers a new account
func (h *AccountHandler) registerAccount(w http.ResponseWriter, r *http.Request) {
	var storeReq storeRequest
	if err := json.NewDecoder(r.Body).Decode(&storeReq); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	validate.RegisterValidation("currency", validCurrency)

	err := validate.Var(storeReq.Currency, "currency")
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, currencyNotSupported, http.StatusBadRequest)
		return
	}

	err = validate.Struct(storeReq)
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	acct := accountRequestFromAccountDomain(storeReq)
	account, err := h.Service.Register(r.Context(), acct)
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(account); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// clean deletes an account by ID
func (h *AccountHandler) clean(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.Logger.Error("no account id found")
		http.Error(w, accountIDRequired, http.StatusBadRequest)
		return
	}

	err := h.Service.Clean(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
