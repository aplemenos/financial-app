package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"financial-app/domain/account"
	"financial-app/register"
)

var (
	accountIDRequired    = "account id required"
	currencyNotSupported = "currency is not supported"
)

type registerHandler struct {
	s register.Service

	logger *zap.SugaredLogger
}

// router sets up all the routes for account service
func (h *registerHandler) router() chi.Router {
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
func (h *registerHandler) loadAccount(w http.ResponseWriter, r *http.Request) {
	id := account.AccountID(chi.URLParam(r, "id"))
	if id == "" {
		h.logger.Error("no account id found")
		http.Error(w, accountIDRequired, http.StatusBadRequest)
		return
	}

	acct, err := h.s.LoadAccount(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(acct); err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// accounts retrieves all the registered accounts
func (h *registerHandler) accounts(w http.ResponseWriter, r *http.Request) {
	accounts := h.s.Accounts(r.Context())

	if err := json.NewEncoder(w).Encode(accounts); err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// validCurrency validates if the given currency is supported
func validCurrency(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return account.IsSupportedCurrency(account.Currency(currency))
	}
	return false
}

// storeRequest
type storeRequest struct {
	Balance  float64 `json:"balance" validate:"required"`
	Currency string  `json:"currency" validate:"currency"`
}

func accountRequestFromAccountDomain(p storeRequest) account.Account {
	return account.Account{
		ID:       account.NextAccountID(), // Generate a new uuid
		Balance:  p.Balance,
		Currency: account.Currency(p.Currency),
	}
}

// registerAccount registers a new account
func (h *registerHandler) registerAccount(w http.ResponseWriter, r *http.Request) {
	var storeReq storeRequest
	if err := json.NewDecoder(r.Body).Decode(&storeReq); err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	validate.RegisterValidation("currency", validCurrency)

	err := validate.Var(storeReq.Currency, "currency")
	if err != nil {
		h.logger.Error(err)
		http.Error(w, currencyNotSupported, http.StatusBadRequest)
		return
	}

	err = validate.Struct(storeReq)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	acct := accountRequestFromAccountDomain(storeReq)
	account, err := h.s.Register(r.Context(), acct)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(account); err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// clean deletes an account by ID
func (h *registerHandler) clean(w http.ResponseWriter, r *http.Request) {
	id := account.AccountID(chi.URLParam(r, "id"))
	if id == "" {
		h.logger.Error("no account id found")
		http.Error(w, accountIDRequired, http.StatusBadRequest)
		return
	}

	err := h.s.Clean(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(
		response{Message: "Successfully Deleted"}); err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
