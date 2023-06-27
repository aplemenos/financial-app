package api

import (
	"financial-app/internals/transaction/rest"

	"github.com/gorilla/mux"
)

func FinancialAppV1() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/transactions", rest.CreateTransaction).Methods("POST")
	r.HandleFunc("/transactions", rest.GetTransactions).Methods("GET")
	r.HandleFunc("/transactions/{id}", rest.GetTransaction).Methods("GET")

	r.HandleFunc("/accounts", rest.GetAccounts).Methods("GET")

	return r
}
