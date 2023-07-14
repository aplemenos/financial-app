package transaction

import (
	"context"
	"errors"
	"financial-app/domain/account"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockTransactionRepository is a mock implementation of TransactionRepository
type mockTransactionRepository struct {
	Transactions map[TransactionID]*Transaction
}

func (m *mockTransactionRepository) Transfer(
	ctx context.Context, txn *Transaction, sacc *account.Account, tacc *account.Account,
) (*Transaction, error) {
	if sacc.Balance >= txn.Amount {
		sacc.Balance -= txn.Amount
		tacc.Balance += txn.Amount
		m.Transactions[txn.ID] = txn
		return txn, nil
	}
	return nil, errors.New("Insufficient balance")
}

func (m *mockTransactionRepository) Find(
	ctx context.Context, id TransactionID,
) (*Transaction, error) {
	if txn, ok := m.Transactions[id]; ok {
		return txn, nil
	}
	return nil, ErrFetchingTransaction(id)
}

func (m *mockTransactionRepository) FindAll(
	ctx context.Context,
) ([]*Transaction, error) {
	transactions := make([]*Transaction, 0, len(m.Transactions))
	for _, txn := range m.Transactions {
		transactions = append(transactions, txn)
	}
	return transactions, nil
}

func (m *mockTransactionRepository) Delete(
	ctx context.Context, id TransactionID,
) error {
	if _, ok := m.Transactions[id]; ok {
		delete(m.Transactions, id)
		return nil
	}
	return ErrDeletingTransaction(id)
}

func TestTransactionRepository_Transfer(t *testing.T) {
	repo := &mockTransactionRepository{Transactions: make(map[TransactionID]*Transaction)}

	sourceID := account.AccountID("1111")
	targetID := account.AccountID("2222")
	amount := 100.0
	currency := account.Currency("USD")
	transactionID := NextTransactionID()

	sourceAccount := &account.Account{
		ID:       sourceID,
		Balance:  200.0,
		Currency: currency,
	}
	targetAccount := &account.Account{
		ID:       targetID,
		Balance:  0.0,
		Currency: currency,
	}

	txn := NewTransaction(transactionID, sourceID, targetID, amount, currency)

	transact, err := repo.Transfer(context.Background(), txn, sourceAccount, targetAccount)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(
		t,
		100.0,
		sourceAccount.Balance,
		"Source account balance should be decreased by the transaction amount",
	)
	assert.Equal(
		t,
		100.0,
		targetAccount.Balance,
		"Target account balance should be increased by the transaction amount",
	)
	assert.Equal(
		t,
		txn,
		transact,
		"Transaction should be stored in the repository",
	)
}

func TestTransactionRepository_TransferInsufficientBalance(t *testing.T) {
	repo := &mockTransactionRepository{Transactions: make(map[TransactionID]*Transaction)}

	sourceID := account.AccountID("1111")
	targetID := account.AccountID("2222")
	amount := 200.0
	currency := account.Currency("USD")
	transactionID := NextTransactionID()

	sourceAccount := &account.Account{
		ID:       sourceID,
		Balance:  100.0,
		Currency: currency,
	}
	targetAccount := &account.Account{
		ID:       targetID,
		Balance:  0.0,
		Currency: currency,
	}

	txn := NewTransaction(transactionID, sourceID, targetID, amount, currency)

	_, err := repo.Transfer(context.Background(), txn, sourceAccount, targetAccount)

	assert.Error(t, err, "Error should not be nil")
	assert.Equal(
		t,
		errors.New("Insufficient balance"),
		err,
		"Error should match ErrInsufficientBalance",
	)
	assert.NotContains(
		t,
		repo.Transactions,
		transactionID,
		"Transaction should not be stored in the repository",
	)
}

func TestTransactionRepository_Find(t *testing.T) {
	repo := &mockTransactionRepository{
		Transactions: map[TransactionID]*Transaction{
			TransactionID("1111"): {ID: TransactionID("1111")},
		},
	}

	id := TransactionID("1111")
	txn, err := repo.Find(context.Background(), id)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, id, txn.ID, "Found transaction ID should match the provided ID")
}

func TestTransactionRepository_Delete(t *testing.T) {
	repo := &mockTransactionRepository{
		Transactions: map[TransactionID]*Transaction{
			TransactionID("1111"): {ID: TransactionID("1111")},
		},
	}

	id := TransactionID("1111")
	err := repo.Delete(context.Background(), id)

	assert.NoError(t, err, "Error should be nil")
	_, exists := repo.Transactions[id]
	assert.False(t, exists, "Transaction should be deleted")
}

func TestTransactionRepository_FindAll(t *testing.T) {
	repo := &mockTransactionRepository{
		Transactions: map[TransactionID]*Transaction{
			TransactionID("1111"): {ID: TransactionID("1111")},
			TransactionID("2222"): {ID: TransactionID("2222")},
			TransactionID("3333"): {ID: TransactionID("3333")},
		},
	}

	transactions, err := repo.FindAll(context.Background())

	assert.NoError(t, err, "Error should be nil")
	assert.Len(t, transactions, 3, "Number of returned transactions should be 3")
	assert.Contains(
		t,
		transactions,
		&Transaction{ID: TransactionID("1111")},
		"Transaction 1 should be present",
	)
	assert.Contains(
		t,
		transactions,
		&Transaction{ID: TransactionID("2222")},
		"Transaction 2 should be present",
	)
	assert.Contains(
		t,
		transactions,
		&Transaction{ID: TransactionID("3333")},
		"Transaction 3 should be present",
	)
}

func TestTransactionRepository_FindAllTransactionsEmpty(t *testing.T) {
	repo := &mockTransactionRepository{
		Transactions: make(map[TransactionID]*Transaction),
	}

	transactions, err := repo.FindAll(context.Background())

	assert.NoError(t, err, "Error should be nil")
	assert.Empty(t, transactions, "Returned transactions should be empty")
}
