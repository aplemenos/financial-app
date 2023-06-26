package main

import (
	"financial-app/internals/transaction/rest"
	"financial-app/mocks"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Generate mock accounts
	mocks.GenerateMockAccounts()

	r := mux.NewRouter()
	r.HandleFunc("/transactions", rest.CreateTransaction).Methods("POST")
	r.HandleFunc("/transactions", rest.GetTransactions).Methods("GET")
	r.HandleFunc("/transactions/{id}", rest.GetTransaction).Methods("GET")

	r.HandleFunc("/accounts", rest.GetAccounts).Methods("GET")

	log.Println("Financial Server started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
