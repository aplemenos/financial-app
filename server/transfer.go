package server

import (
	"encoding/json"
	"financial-app/domain/account"
	"financial-app/domain/transaction"
	"financial-app/transfer"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

var (
	transactionIDRequired = "transaction id required"
)

type transferHandler struct {
	s transfer.Service

	logger *zap.SugaredLogger
}

// router sets up all the routes for tranfer service
func (h *transferHandler) router() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.transfer)
	r.Get("/", h.transactions)
	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.loadTransaction)
		r.Delete("/", h.clean)
	})

	return r
}

// loadTransaction retrieves a transaction by ID
func (h *transferHandler) loadTransaction(w http.ResponseWriter, r *http.Request) {
	id := transaction.TransactionID(chi.URLParam(r, "id"))
	if id == "" {
		h.logger.Error("no transaction id found")
		http.Error(w, transactionIDRequired, http.StatusBadRequest)
		return
	}

	txn, err := h.s.LoadTransaction(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(txn); err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// transactions retrieves all the completed transactions
func (h *transferHandler) transactions(w http.ResponseWriter, r *http.Request) {
	txns := h.s.Transactions(r.Context())

	if err := json.NewEncoder(w).Encode(txns); err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// transferRequest
type transferRequest struct {
	SourceAccountID string  `json:"source_account_id" validate:"required,uuid"`
	TargetAccountID string  `json:"target_account_id" validate:"required,uuid"`
	Amount          float64 `json:"amount" validate:"required,numeric,gt=0"`
	Currency        string  `json:"currency" validate:"currency"`
}

func transferRequestFromTransactionDomain(p transferRequest) transaction.Transaction {
	return transaction.Transaction{
		ID:              transaction.NextTransactionID(), // Generate a new uuid
		SourceAccountID: account.AccountID(p.SourceAccountID),
		TargetAccountID: account.AccountID(p.TargetAccountID),
		Amount:          p.Amount,
		Currency:        account.Currency(p.Currency),
	}
}

// transfer performs a new transaction from source to target account
func (h *transferHandler) transfer(w http.ResponseWriter, r *http.Request) {
	var transferReq transferRequest

	if err := json.NewDecoder(r.Body).Decode(&transferReq); err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v := validator.New()
	v.RegisterValidation("currency", validCurrency)

	err := v.Var(transferReq.Currency, "currency")
	if err != nil {
		h.logger.Error(err)
		http.Error(w, currencyNotSupported, http.StatusBadRequest)
		return
	}

	err = v.Struct(transferReq)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = v.VarWithValue(transferReq.SourceAccountID, transferReq.TargetAccountID, "necsfield")
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	txn := transferRequestFromTransactionDomain(transferReq)
	transact, err := h.s.Transfer(r.Context(), txn)
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
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// clean deletes a transaction by ID
func (h *transferHandler) clean(w http.ResponseWriter, r *http.Request) {
	id := transaction.TransactionID(chi.URLParam(r, "id"))
	if id == "" {
		h.logger.Error("no transaction id found")
		http.Error(w, transactionIDRequired, http.StatusBadRequest)
		return
	}

	err := h.s.Clean(r.Context(), id)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(response{Message: "Successfully Deleted"}); err != nil {
		h.logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
