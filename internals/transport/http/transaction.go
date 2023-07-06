package http

import (
	"context"
	"encoding/json"
	"errors"
	"financial-app/internals/service"
	"financial-app/pkg/account"
	"financial-app/pkg/transaction"
	"net/http"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"

	log "github.com/sirupsen/logrus"

	"github.com/go-playground/validator/v10"
)

var (
	AccountIDRequired     = "account id required"
	TransactionIDRequired = "transaction id required"
)

type TransactionService interface {
	GetAccount(ctx context.Context, ID string) (account.Account, error)
	PostAccount(ctx context.Context, acct account.Account) (account.Account, error)
	DeleteAccount(ctx context.Context, ID string) error

	GetTransaction(ctx context.Context, ID string) (transaction.Transaction, error)
	DeleteTransaction(ctx context.Context, ID string) error

	Transfer(ctx context.Context, txn transaction.Transaction) (transaction.Transaction, error)

	AliveCheck(ctx context.Context) error
}

// GetAccount - retrieve an account by ID
func (h *Handler) GetAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		log.Error("no account id found")
		http.Error(w, AccountIDRequired, http.StatusBadRequest)
		return
	}

	txn, err := h.TransactionService.GetAccount(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrFetchingAccount) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(txn); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// PostAccountRequest
type PostAccountRequest struct {
	Balance  float64 `json:"balance" validate:"required"`
	Currency string  `json:"currency" validate:"required"`
}

func accountFromPostAccountRequest(p PostAccountRequest) account.Account {
	return account.Account{
		ID:       uuid.NewV4().String(), // Generate a new uuid
		Balance:  p.Balance,
		Currency: p.Currency,
	}
}

// PostAccount - adds a new account
func (h *Handler) PostAccount(w http.ResponseWriter, r *http.Request) {
	var postAcctReq PostAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&postAcctReq); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	err := validate.Struct(postAcctReq)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	acct := accountFromPostAccountRequest(postAcctReq)
	acct, err = h.TransactionService.PostAccount(r.Context(), acct)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(acct); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// DeleteAccount - deletes an account by ID
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		log.Error("no account id found")
		http.Error(w, AccountIDRequired, http.StatusBadRequest)
		return
	}

	err := h.TransactionService.DeleteAccount(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(Response{Message: "Successfully Deleted"}); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetTransaction - retrieve a transaction by ID
func (h *Handler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		log.Error("no transaction id found")
		http.Error(w, TransactionIDRequired, http.StatusBadRequest)
		return
	}

	txn, err := h.TransactionService.GetTransaction(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrFetchingTransaction) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(txn); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// PostTransactionRequest
type PostTransactionRequest struct {
	SourceAccountID string  `json:"source_account_id" validate:"required"`
	TargetAccountID string  `json:"target_account_id" validate:"required"`
	Amount          float64 `json:"amount" validate:"required,numeric,gt=0"`
	Currency        string  `json:"currency" validate:"required"`
}

func transactionFromPostTransactionRequest(p PostTransactionRequest) transaction.Transaction {
	return transaction.Transaction{
		ID:              uuid.NewV4().String(), // Generate a new uuid
		SourceAccountID: p.SourceAccountID,
		TargetAccountID: p.TargetAccountID,
		Amount:          p.Amount,
		Currency:        p.Currency,
	}
}

// Transfer - performs a new transaction from source to target account
func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	var postTxnReq PostTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&postTxnReq); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	v := validator.New()

	err := v.Struct(postTxnReq)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = v.VarWithValue(postTxnReq.SourceAccountID, postTxnReq.TargetAccountID, "necsfield")
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	txn := transactionFromPostTransactionRequest(postTxnReq)
	txn, err = h.TransactionService.Transfer(r.Context(), txn)
	if err != nil {
		if errors.Is(err, service.ErrNoAccountFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrInsufficientBalance) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(txn); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// DeleteTransaction - deletes a transaction by ID
func (h *Handler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txnID := vars["id"]

	if txnID == "" {
		log.Error("no transaction id found")
		http.Error(w, TransactionIDRequired, http.StatusBadRequest)
		return
	}

	err := h.TransactionService.DeleteTransaction(r.Context(), txnID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(Response{Message: "Successfully Deleted"}); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) AliveCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.TransactionService.AliveCheck(r.Context()); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(Response{Message: "I am Alive!"}); err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}