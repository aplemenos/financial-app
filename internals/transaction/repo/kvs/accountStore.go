package kvs

import (
	"errors"
	"financial-app/pkg/models"
	"sync"
)

var (
	asInstance   sync.Once
	accountStore *AccountStore
)

// AccountStore represents a key-value store for accounts
type AccountStore struct {
	accounts map[string]models.Account
	lock     sync.RWMutex
}

// NewAccountStore creates a single instance of AccountStore
func NewAccountStore() *AccountStore {
	asInstance.Do(func() { // <-- atomic, does not allow repeating
		accountStore = new(AccountStore) // <-- thread safe
		accountStore.accounts = make(map[string]models.Account)
	})

	return accountStore
}

// Set sets a new account for a given key (uuid)
func (as *AccountStore) Set(key string, account models.Account) {
	as.lock.Lock()
	defer as.lock.Unlock()

	as.accounts[key] = account
}

// Get retrieves the account for a given key (uuid)
func (as *AccountStore) Get(key string) (*models.Account, error) {
	as.lock.RLock()
	defer as.lock.RUnlock()

	account, ok := as.accounts[key]
	if !ok {
		return nil, errors.New("account not found")
	}

	return &account, nil
}

func (as *AccountStore) GetAll() []models.Account {
	as.lock.RLock()
	defer as.lock.RUnlock()

	accounts := []models.Account{}
	for _, account := range as.accounts {
		accounts = append(accounts, account)
	}
	return accounts
}

// Delete deletes an account pair from the account store
func (as *AccountStore) Delete(key string) {
	as.lock.Lock()
	defer as.lock.Unlock()

	delete(as.accounts, key)
}
