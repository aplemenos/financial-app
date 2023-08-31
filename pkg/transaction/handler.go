package transaction

import (
	"encoding/json"
	"financial-app/pkg/account"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var (
	transactionIDRequired = "transaction id required"
	currencyNotSupported  = "currency is not supported"
	accountsNotSame       = "accounts cannot be the same"
)

type TransactionHandler struct {
	Service Service

	Logger *zap.SugaredLogger
}

// router sets up all the routes for tranfer service
func (h *TransactionHandler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.transfer)
	r.Get("/", h.loadAll)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.load)
		r.Delete("/", h.clean)
	})

	return r
}

// load retrieves a transaction by ID
func (h *TransactionHandler) load(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.Logger.Error("no transaction id found")
		http.Error(w, transactionIDRequired, http.StatusBadRequest)
		return
	}

	txn, err := h.Service.Load(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(txn); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// loadAll retrieves all the completed loadAll
func (h *TransactionHandler) loadAll(w http.ResponseWriter, r *http.Request) {
	txns := h.Service.LoadAll(r.Context())

	if err := json.NewEncoder(w).Encode(txns); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// transferRequest
type transactionRequest struct {
	SourceAccountID string  `json:"source_account_id" validate:"required,uuid"`
	TargetAccountID string  `json:"target_account_id" validate:"required,uuid"`
	Amount          float64 `json:"amount" validate:"required,numeric,gt=0"`
	Currency        string  `json:"currency" validate:"currency"`
}

func transactionRequestFromTransactionDomain(p transactionRequest) Transaction {
	return Transaction{
		ID:              nextTransactionID(), // Generate a new uuid
		SourceAccountID: p.SourceAccountID,
		TargetAccountID: p.TargetAccountID,
		Amount:          p.Amount,
		Currency:        p.Currency,
	}
}

// validCurrency validates if the given currency is supported
func validCurrency(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return account.IsSupportedCurrency(currency)
	}
	return false
}

// transfer performs a new transaction from source to target account
func (h *TransactionHandler) transfer(w http.ResponseWriter, r *http.Request) {
	var transactionReq transactionRequest

	if err := json.NewDecoder(r.Body).Decode(&transactionReq); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v := validator.New()
	v.RegisterValidation("currency", validCurrency)

	err := v.Var(transactionReq.Currency, "currency")
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, currencyNotSupported, http.StatusBadRequest)
		return
	}

	err = v.Struct(transactionReq)
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = v.VarWithValue(transactionReq.SourceAccountID,
		transactionReq.TargetAccountID, "necsfield")
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, accountsNotSame, http.StatusConflict)
		return
	}

	txn := transactionRequestFromTransactionDomain(transactionReq)
	transact, err := h.Service.Transfer(r.Context(), txn)
	if err != nil {
		noSourceAccountFound := account.ErrFetchingAccount(txn.SourceAccountID).Error()
		noTargetAccountFound := account.ErrFetchingAccount(txn.TargetAccountID).Error()
		if err.Error() == noSourceAccountFound ||
			err.Error() == noTargetAccountFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Insufficient balance error
		if strings.Contains(err.Error(), "insufficient") {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(transact); err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// clean deletes a transaction by ID
func (h *TransactionHandler) clean(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.Logger.Error("no transaction id found")
		http.Error(w, transactionIDRequired, http.StatusBadRequest)
		return
	}

	err := h.Service.Clean(r.Context(), id)
	if err != nil {
		h.Logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
