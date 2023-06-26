package rest

import (
	"encoding/json"
	"financial-app/internals/transaction/repo/kvs"
	"financial-app/internals/transaction/service"
	"financial-app/pkg/models"
	"net/http"

	"github.com/gorilla/mux"
)

func CreateTransaction(w http.ResponseWriter, r *http.Request) {
	transaction := models.Transaction{}
	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get a single instance of account and transaction stores
	t := new(service.Transaction)
	t.Accounts = kvs.NewAccountStore()
	t.Transactions = kvs.NewTransactionStore()
	// Perform the transaction
	transactionID, err := t.CreateTransanction(transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(transactionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetTransaction(w http.ResponseWriter, r *http.Request) {
	id, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, "the parameter id does not exist", http.StatusBadRequest)
	}

	// Get a single instance of account and transaction stores
	t := new(service.Transaction)
	t.Accounts = kvs.NewAccountStore()
	t.Transactions = kvs.NewTransactionStore()
	// Get a transaction by ID
	transaction, err := t.GetTransanction(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(transaction)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func GetTransactions(w http.ResponseWriter, r *http.Request) {
	// Get a single instance of account and transaction stores
	t := new(service.Transaction)
	t.Accounts = kvs.NewAccountStore()
	t.Transactions = kvs.NewTransactionStore()
	// Get all transactions
	transactions := t.GetTransactions()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(transactions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
