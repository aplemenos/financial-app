package transactions

import (
	"context"
	"financial-app/pkg/accounts"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAccountRepository struct {
	Accounts map[string]*accounts.Account
}

func (m *mockAccountRepository) Find(
	ctx context.Context, id string,
) (*accounts.Account, error) {
	if acct, ok := m.Accounts[id]; ok {
		return acct, nil
	}
	return nil, accounts.ErrFetchingAccount(id)
}

func (m *mockAccountRepository) FindByIDs(
	ctx context.Context, ids []string,
) (map[string]*accounts.Account, error) {
	accounts := make(map[string]*accounts.Account)
	for _, id := range ids {
		if acct, ok := m.Accounts[id]; ok {
			accounts[id] = acct
		}
	}
	return accounts, nil
}

func (m *mockAccountRepository) Store(
	ctx context.Context, acct *accounts.Account,
) (*accounts.Account, error) {
	m.Accounts[acct.ID] = acct
	return acct, nil
}

func (m *mockAccountRepository) FindAll(
	ctx context.Context,
) []*accounts.Account {
	accounts := make([]*accounts.Account, 0, len(m.Accounts))
	for _, txn := range m.Accounts {
		accounts = append(accounts, txn)
	}
	return accounts
}

func (m *mockAccountRepository) Delete(
	ctx context.Context, id string,
) error {
	if _, ok := m.Accounts[id]; ok {
		delete(m.Accounts, id)
		return nil
	}
	return accounts.ErrDeletingAccount(id)
}

type mockTransactionRepository struct {
	Transactions map[string]*Transaction
}

func (m *mockTransactionRepository) Transfer(
	ctx context.Context, txn *Transaction, sacc *accounts.Account, tacc *accounts.Account,
) (*Transaction, error) {
	if sacc.Balance >= txn.Amount {
		m.Transactions[txn.ID] = txn
		return txn, nil
	}
	return nil,
		errInsufficientBalance(sacc.Balance, txn.Amount, sacc.ID)
}

func (m *mockTransactionRepository) Find(
	ctx context.Context, id string,
) (*Transaction, error) {
	if txn, ok := m.Transactions[id]; ok {
		return txn, nil
	}
	return nil, ErrFetchingTransaction(id)
}

func (m *mockTransactionRepository) FindAll(
	ctx context.Context,
) []*Transaction {
	transactions := make([]*Transaction, 0, len(m.Transactions))
	for _, txn := range m.Transactions {
		transactions = append(transactions, txn)
	}
	return transactions
}

func (m *mockTransactionRepository) Delete(
	ctx context.Context, id string,
) error {
	if _, ok := m.Transactions[id]; ok {
		delete(m.Transactions, id)
		return nil
	}
	return ErrDeletingTransaction(id)
}

func TestService_LoadTransaction(t *testing.T) {
	transactionID := "1111"
	sourceAccountID := "2222"
	targetAccountID := "3333"

	expectedTransaction := Transaction{
		ID:              transactionID,
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          100.0,
		Currency:        "USD",
	}

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: map[string]*Transaction{
			transactionID: &expectedTransaction,
		},
	}

	service := NewService(nil, mockTransactionRepository)

	loadedTransaction, err := service.Load(context.Background(), transactionID)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(
		t,
		expectedTransaction,
		loadedTransaction,
		"Loaded transaction should match the mock transaction",
	)
}

func TestService_TransferHappyPath(t *testing.T) {
	sourceAccountID := "2222"
	targetAccountID := "3333"

	mockSourceAccount := accounts.Account{
		ID:       sourceAccountID,
		Balance:  200.0,
		Currency: "USD",
	}

	mockTargetAccount := accounts.Account{
		ID:       targetAccountID,
		Balance:  0.0,
		Currency: "USD",
	}

	mockAccounts := map[string]*accounts.Account{
		sourceAccountID: &mockSourceAccount,
		targetAccountID: &mockTargetAccount,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccounts,
	}

	expectedTransaction := Transaction{
		ID:              "1111",
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          100.0,
		Currency:        "USD",
	}

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: make(map[string]*Transaction),
	}

	service := NewService(mockAccountRepository, mockTransactionRepository)

	transferedTransaction, err := service.Transfer(context.Background(), expectedTransaction)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(
		t,
		expectedTransaction,
		transferedTransaction,
		"Transfered transaction should match the mock transaction",
	)
	assert.Equal(
		t,
		100.0,
		mockSourceAccount.Balance,
		"Source account balance should be decreased by the transaction amount",
	)
	assert.Equal(
		t,
		100.0,
		mockTargetAccount.Balance,
		"Target account balance should be increased by the transaction amount",
	)
}

func TestService_TransferInsufficientBalance(t *testing.T) {
	sourceAccountID := "2222"
	targetAccountID := "3333"

	mockSourceAccount := accounts.Account{
		ID:       sourceAccountID,
		Balance:  100.0,
		Currency: "USD",
	}

	mockTargetAccount := accounts.Account{
		ID:       targetAccountID,
		Balance:  0.0,
		Currency: "USD",
	}

	mockAccounts := map[string]*accounts.Account{
		sourceAccountID: &mockSourceAccount,
		targetAccountID: &mockTargetAccount,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccounts,
	}

	mockTransaction := Transaction{
		ID:              "1111",
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          200.0,
		Currency:        "USD",
	}

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: make(map[string]*Transaction),
	}

	service := NewService(mockAccountRepository, mockTransactionRepository)

	_, err := service.Transfer(context.Background(), mockTransaction)

	assert.Error(t, err, "Error should not be nil")
	assert.Equal(
		t,
		errInsufficientBalance(
			mockSourceAccount.Balance,
			mockTransaction.Amount,
			mockSourceAccount.ID,
		),
		err,
		"Error should match the expected error",
	)
	assert.Equal(
		t,
		100.0,
		mockSourceAccount.Balance,
		"Source account balance should not be changed",
	)
	assert.Equal(
		t,
		0.0,
		mockTargetAccount.Balance,
		"Target account balance should not be changed",
	)
}

func TestService_Transactions(t *testing.T) {
	expectedTransaction1 := Transaction{
		ID:              "1111",
		SourceAccountID: "2222",
		TargetAccountID: "3333",
		Amount:          100.0,
		Currency:        "USD",
	}

	expectedTransaction2 := Transaction{
		ID:              "4444",
		SourceAccountID: "5555",
		TargetAccountID: "6666",
		Amount:          200.0,
		Currency:        "USD",
	}

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: map[string]*Transaction{
			expectedTransaction1.ID: &expectedTransaction1,
			expectedTransaction2.ID: &expectedTransaction2,
		},
	}

	service := NewService(nil, mockTransactionRepository)

	transactions := service.LoadAll(context.Background())

	assert.Len(t, transactions, 2, "Number of returned transactions should be 2")
	assert.Contains(
		t,
		transactions,
		expectedTransaction1,
		"Transaction 1 should be present",
	)
	assert.Contains(
		t,
		transactions,
		expectedTransaction2,
		"Transaction 2 should be present",
	)
}

func TestService_Clean(t *testing.T) {
	mockTransactionID := "transaction-123"

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: map[string]*Transaction{
			mockTransactionID: {},
		},
	}

	service := NewService(nil, mockTransactionRepository)

	err := service.Clean(context.Background(), mockTransactionID)

	assert.NoError(t, err, "Error should be nil")
	_, exists := mockTransactionRepository.Transactions[mockTransactionID]
	assert.False(t, exists, "Transaction should be deleted")
}
