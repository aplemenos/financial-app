package rest

import (
	"encoding/json"
	"financial-app/internals/transaction/repo/kvs"
	"financial-app/internals/transaction/service"
	"net/http"
)

func GetAccounts(w http.ResponseWriter, r *http.Request) {
	// Get a single instance of account store
	a := new(service.Account)
	a.Accounts = kvs.NewAccountStore()
	// Get all accounts
	accounts := a.GetAccounts()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	err := json.NewEncoder(w).Encode(accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
