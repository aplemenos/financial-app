package kvs

import (
	"errors"
	"financial-app/pkg/models"
	"sync"
)

var (
	tsInstance       sync.Once
	transactionStore *TransactionStore
)

// TransactionStore represents a key-value store for transactions
type TransactionStore struct {
	transactions map[string]models.Transaction
	lock         sync.RWMutex
}

// NewTransactionStore creates a single instance of TransactionStore
func NewTransactionStore() *TransactionStore {
	tsInstance.Do(func() { // <-- atomic, does not allow repeating
		transactionStore = new(TransactionStore) // <-- thread safe
		transactionStore.transactions = make(map[string]models.Transaction)
	})

	return transactionStore
}

// Set sets a new transaction for a given key (uuid)
func (ts *TransactionStore) Set(key string, transaction models.Transaction) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	ts.transactions[key] = transaction
}

// Get retrieves the transaction for a given key (uuid)
func (ts *TransactionStore) Get(key string) (*models.Transaction, error) {
	ts.lock.RLock()
	defer ts.lock.RUnlock()

	transaction, ok := ts.transactions[key]
	if !ok {
		return nil, errors.New("transaction not found")
	}

	return &transaction, nil
}

func (ts *TransactionStore) GetAll() []models.Transaction {
	ts.lock.RLock()
	defer ts.lock.RUnlock()

	transactions := []models.Transaction{}
	for _, transaction := range ts.transactions {
		transactions = append(transactions, transaction)
	}
	return transactions
}

// Delete deletes a transaction pair from the transaction store
func (ts *TransactionStore) Delete(key string) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	delete(ts.transactions, key)
}
