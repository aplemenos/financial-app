package transfer

import (
	"context"
	"financial-app/domain/account"
	"financial-app/domain/transaction"
	"financial-app/multiplelock"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAccountRepository struct {
	Accounts map[account.AccountID]*account.Account
}

func (m *mockAccountRepository) Find(
	ctx context.Context, id account.AccountID,
) (*account.Account, error) {
	if acct, ok := m.Accounts[id]; ok {
		return acct, nil
	}
	return nil, account.ErrFetchingAccount(id)
}

func (m *mockAccountRepository) FindByIDs(
	ctx context.Context, ids []account.AccountID,
) (map[account.AccountID]*account.Account, error) {
	accounts := make(map[account.AccountID]*account.Account)
	for _, id := range ids {
		if acct, ok := m.Accounts[id]; ok {
			accounts[id] = acct
		}
	}
	return accounts, nil
}

func (m *mockAccountRepository) Store(
	ctx context.Context, acct *account.Account,
) (*account.Account, error) {
	m.Accounts[acct.ID] = acct
	return acct, nil
}

func (m *mockAccountRepository) FindAll(
	ctx context.Context,
) []*account.Account {
	accounts := make([]*account.Account, 0, len(m.Accounts))
	for _, txn := range m.Accounts {
		accounts = append(accounts, txn)
	}
	return accounts
}

func (m *mockAccountRepository) Delete(
	ctx context.Context, id account.AccountID,
) error {
	if _, ok := m.Accounts[id]; ok {
		delete(m.Accounts, id)
		return nil
	}
	return account.ErrDeletingAccount(id)
}

type mockTransactionRepository struct {
	Transactions map[transaction.TransactionID]*transaction.Transaction
}

func (m *mockTransactionRepository) Transfer(
	ctx context.Context, txn *transaction.Transaction, sacc *account.Account, tacc *account.Account,
) (*transaction.Transaction, error) {
	if sacc.Balance >= txn.Amount {
		m.Transactions[txn.ID] = txn
		return txn, nil
	}
	return nil,
		errInsufficientBalance(sacc.Balance, txn.Amount, sacc.ID)
}

func (m *mockTransactionRepository) Find(
	ctx context.Context, id transaction.TransactionID,
) (*transaction.Transaction, error) {
	if txn, ok := m.Transactions[id]; ok {
		return txn, nil
	}
	return nil, transaction.ErrFetchingTransaction(id)
}

func (m *mockTransactionRepository) FindAll(
	ctx context.Context,
) []*transaction.Transaction {
	transactions := make([]*transaction.Transaction, 0, len(m.Transactions))
	for _, txn := range m.Transactions {
		transactions = append(transactions, txn)
	}
	return transactions
}

func (m *mockTransactionRepository) Delete(
	ctx context.Context, id transaction.TransactionID,
) error {
	if _, ok := m.Transactions[id]; ok {
		delete(m.Transactions, id)
		return nil
	}
	return transaction.ErrDeletingTransaction(id)
}

func (m *mockTransactionRepository) Ping(ctx context.Context) error {
	return nil
}

func TestService_LoadTransaction(t *testing.T) {
	transactionID := transaction.TransactionID("1111")
	sourceAccountID := account.AccountID("2222")
	targetAccountID := account.AccountID("3333")

	mockTransaction := &transaction.Transaction{
		ID:              transactionID,
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          100.0,
		Currency:        account.Currency("USD"),
	}

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: map[transaction.TransactionID]*transaction.Transaction{
			transactionID: mockTransaction,
		},
	}

	service := NewService(nil, mockTransactionRepository, nil)

	loadedTransaction, err := service.LoadTransaction(context.Background(), transactionID)

	expectedTransaction := assemble(mockTransaction)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(
		t,
		expectedTransaction,
		loadedTransaction,
		"Loaded transaction should match the mock transaction",
	)
}

func TestService_TransferHappyPath(t *testing.T) {
	sourceAccountID := account.AccountID("2222")
	targetAccountID := account.AccountID("3333")

	mockSourceAccount := &account.Account{
		ID:       sourceAccountID,
		Balance:  200.0,
		Currency: account.Currency("USD"),
	}

	mockTargetAccount := &account.Account{
		ID:       targetAccountID,
		Balance:  0.0,
		Currency: account.Currency("USD"),
	}

	mockAccounts := map[account.AccountID]*account.Account{
		sourceAccountID: mockSourceAccount,
		targetAccountID: mockTargetAccount,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccounts,
	}

	mockTransaction := &transaction.Transaction{
		ID:              transaction.TransactionID("1111"),
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          100.0,
		Currency:        account.Currency("USD"),
	}

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: make(map[transaction.TransactionID]*transaction.Transaction),
	}

	mlock := multiplelock.NewMultipleLock()

	service := NewService(mockAccountRepository, mockTransactionRepository, mlock)

	transferedTransaction, err := service.Transfer(context.Background(), *mockTransaction)

	expectedTransaction := assemble(mockTransaction)

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
	sourceAccountID := account.AccountID("2222")
	targetAccountID := account.AccountID("3333")

	mockSourceAccount := &account.Account{
		ID:       sourceAccountID,
		Balance:  100.0,
		Currency: account.Currency("USD"),
	}

	mockTargetAccount := &account.Account{
		ID:       targetAccountID,
		Balance:  0.0,
		Currency: account.Currency("USD"),
	}

	mockAccounts := map[account.AccountID]*account.Account{
		sourceAccountID: mockSourceAccount,
		targetAccountID: mockTargetAccount,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccounts,
	}

	mockTransaction := &transaction.Transaction{
		ID:              transaction.TransactionID("1111"),
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          200.0,
		Currency:        account.Currency("USD"),
	}

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: make(map[transaction.TransactionID]*transaction.Transaction),
	}

	mlock := multiplelock.NewMultipleLock()

	service := NewService(mockAccountRepository, mockTransactionRepository, mlock)

	_, err := service.Transfer(context.Background(), *mockTransaction)

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
	mockTransaction1 := &transaction.Transaction{
		ID:              transaction.TransactionID("1111"),
		SourceAccountID: account.AccountID("2222"),
		TargetAccountID: account.AccountID("3333"),
		Amount:          100.0,
		Currency:        account.Currency("USD"),
	}

	mockTransaction2 := &transaction.Transaction{
		ID:              transaction.TransactionID("4444"),
		SourceAccountID: account.AccountID("5555"),
		TargetAccountID: account.AccountID("6666"),
		Amount:          200.0,
		Currency:        account.Currency("USD"),
	}

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: map[transaction.TransactionID]*transaction.Transaction{
			mockTransaction1.ID: mockTransaction1,
			mockTransaction2.ID: mockTransaction2,
		},
	}

	service := NewService(nil, mockTransactionRepository, nil)

	transactions := service.Transactions(context.Background())

	assert.Len(t, transactions, 2, "Number of returned transactions should be 2")
	assert.Contains(
		t,
		transactions,
		assemble(mockTransaction1),
		"Transaction 1 should be present",
	)
	assert.Contains(
		t,
		transactions,
		assemble(mockTransaction2),
		"Transaction 2 should be present",
	)
}

func TestService_Clean(t *testing.T) {
	mockTransactionID := transaction.TransactionID("transaction-123")

	mockTransactionRepository := &mockTransactionRepository{
		Transactions: map[transaction.TransactionID]*transaction.Transaction{
			mockTransactionID: {},
		},
	}

	service := NewService(nil, mockTransactionRepository, nil)

	err := service.Clean(context.Background(), mockTransactionID)

	assert.NoError(t, err, "Error should be nil")
	_, exists := mockTransactionRepository.Transactions[mockTransactionID]
	assert.False(t, exists, "Transaction should be deleted")
}
