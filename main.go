package main

import (
	"financial-app/api"
	"financial-app/mocks"
	"log"
	"net/http"
)

func main() {
	// Generate mock accounts
	mocks.GenerateMockAccounts()

	r := api.FinancialAppV1()

	log.Println("Financial Server started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
