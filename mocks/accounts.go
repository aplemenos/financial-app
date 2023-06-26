package mocks

import (
	"financial-app/internals/transaction/repo/kvs"
	"financial-app/pkg/models"
	"time"
)

func GenerateMockAccounts() {
	// Create two mock account data
	account1 := models.Account{
		ID:        "54462360-67e2-47e6-9962-21ba7ec7f141",
		Balance:   1000.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}

	account2 := models.Account{
		ID:        "61b72db1-c001-4bf1-9a72-b2d6c6c8d8bd",
		Balance:   200.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}

	// Store them in the store
	a := kvs.NewAccountStore()
	a.Set(account1.ID, account1)
	a.Set(account2.ID, account2)
}
