package service

import (
	"financial-app/internals/transaction/repo/kvs"
	"financial-app/pkg/models"
	"log"
)

type Account struct {
	Accounts *kvs.AccountStore
}

// NewAccount creates a new instance of Account
func NewAccount(accountStore *kvs.AccountStore) *Account {
	return &Account{
		Accounts: accountStore,
	}
}

func (a *Account) GetAccounts() []models.Account {
	log.Println("Get all accounts")

	return a.Accounts.GetAll()
}
