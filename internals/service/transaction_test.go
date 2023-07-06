package service

import (
	"context"
	"financial-app/pkg/account"
	"financial-app/pkg/transaction"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransactionStore - is a mock implementation of the TransactionStore interface
type MockTransactionStore struct {
	mock.Mock
}

// GetAccount - is a mock implementation of GetAccount method
func (m *MockTransactionStore) GetAccount(
	ctx context.Context, ID string,
) (account.Account, error) {
	args := m.Called(ctx, ID)
	return args.Get(0).(account.Account), args.Error(1)
}

// PostAccount - is a mock implementation of PostAccount method
func (m *MockTransactionStore) PostAccount(
	ctx context.Context, acct account.Account,
) (account.Account, error) {
	args := m.Called(ctx, acct)
	return args.Get(0).(account.Account), args.Error(1)
}

// DeleteAccount - is a mock implementation of DeleteAccount method
func (m *MockTransactionStore) DeleteAccount(ctx context.Context, ID string) error {
	args := m.Called(ctx, ID)
	return args.Error(0)
}

// GetTransaction - is a mock implementation of GetTransaction method
func (m *MockTransactionStore) GetTransaction(
	ctx context.Context, ID string,
) (transaction.Transaction, error) {
	args := m.Called(ctx, ID)
	return args.Get(0).(transaction.Transaction), args.Error(1)
}

// Transfer - is a mock implementation of Transfer method
func (m *MockTransactionStore) Transfer(
	ctx context.Context,
	txn transaction.Transaction,
	sacc account.Account,
	tacc account.Account,
) (transaction.Transaction, error) {
	args := m.Called(ctx, txn, sacc, tacc)
	return args.Get(0).(transaction.Transaction), args.Error(1)
}

// DeleteTransaction - is a mock implementation of DeleteTransaction method
func (m *MockTransactionStore) DeleteTransaction(ctx context.Context, ID string) error {
	args := m.Called(ctx, ID)
	return args.Error(0)
}

// Ping - is the mock implementation for Ping.
func (m *MockTransactionStore) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestTransferHappyPath(t *testing.T) {
	// Create a mock transaction store
	mockStore := new(MockTransactionStore)

	// Create the transaction service with the mock store
	service := NewService(mockStore)

	// Set up the mock data for the test case
	sourceAccountID := "1111"
	targetAccountID := "2222"

	// AC 1: Happy path for money transfer between two accounts
	mockSourceAccount := account.Account{
		ID:       sourceAccountID,
		Balance:  500.00,
		Currency: "EUR",
	}

	mockTargetAccount := account.Account{
		ID:       targetAccountID,
		Balance:  200.00,
		Currency: "EUR",
	}

	mockTransaction := transaction.Transaction{
		ID:              "1234",
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          100.00,
		Currency:        "EUR",
	}

	expSourceAccount := account.Account{
		ID:       sourceAccountID,
		Balance:  400.00,
		Currency: "EUR",
	}

	expTargetAccount := account.Account{
		ID:       targetAccountID,
		Balance:  300.00,
		Currency: "EUR",
	}

	// Set up the expected behavior of the mock store
	mockStore.On("GetAccount", mock.Anything, sourceAccountID).Return(mockSourceAccount, nil)
	mockStore.On("GetAccount", mock.Anything, targetAccountID).Return(mockTargetAccount, nil)
	mockStore.On("Transfer", mock.Anything, mockTransaction, expSourceAccount,
		expTargetAccount).Return(mockTransaction, nil)

	// Call the method being tested
	result, err := service.Transfer(context.Background(), mockTransaction)

	// Assert that the expected methods were called on the mock store
	mockStore.AssertExpectations(t)

	// Assert the result and error
	assert.NoError(t, err)
	assert.Equal(t, mockTransaction.ID, result.ID)
	assert.Equal(t, mockTransaction.SourceAccountID, result.SourceAccountID)
	assert.Equal(t, mockTransaction.TargetAccountID, result.TargetAccountID)
	assert.Equal(t, mockTransaction.Amount, result.Amount)
	assert.Equal(t, mockTransaction.Currency, result.Currency)
}

func TestTransferInsufficientBalance(t *testing.T) {
	// Create a mock transaction store
	mockStore := new(MockTransactionStore)

	// Create the transaction service with the mock store
	service := NewService(mockStore)

	// Set up the mock data for the test case
	sourceAccountID := "1111"
	targetAccountID := "2222"

	// AC 2: Insufficient balance
	mockSourceAccount := account.Account{
		ID:       sourceAccountID,
		Balance:  500.00,
		Currency: "EUR",
	}

	mockTargetAccount := account.Account{
		ID:       targetAccountID,
		Balance:  200.00,
		Currency: "EUR",
	}

	mockTransaction := transaction.Transaction{
		ID:              "1234",
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          600.00,
		Currency:        "EUR",
	}

	// Set up the expected behavior of the mock store
	mockStore.On("GetAccount", mock.Anything, sourceAccountID).Return(mockSourceAccount, nil)
	mockStore.On("GetAccount", mock.Anything, targetAccountID).Return(mockTargetAccount, nil)

	// Call the method being tested
	result, err := service.Transfer(context.Background(), mockTransaction)

	// Assert that the expected methods were called on the mock store
	mockStore.AssertExpectations(t)

	// Assert the result and error
	assert.Error(t, err)
	assert.EqualError(t, err, ErrInsufficientBalance.Error())

	assert.Equal(t, "", result.ID)
	assert.Equal(t, "", result.SourceAccountID)
	assert.Equal(t, "", result.TargetAccountID)
	assert.Equal(t, 0.0, result.Amount)
	assert.Equal(t, "", result.Currency)
}

func TestTransferOneAccountNotFound(t *testing.T) {
	// Create a mock transaction store
	mockStore := new(MockTransactionStore)

	// Create the transaction service with the mock store
	service := NewService(mockStore)

	// Set up the mock data for the test case
	sourceAccountID := "1111"
	targetAccountID := "2222"

	// AC 3: One or more of the accounts does not exist
	mockSourceAccount := account.Account{
		ID:       sourceAccountID,
		Balance:  500.00,
		Currency: "EUR",
	}

	mockTransaction := transaction.Transaction{
		ID:              "1234",
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          100.00,
		Currency:        "EUR",
	}

	// Set up the expected behavior of the mock store
	mockStore.On("GetAccount", mock.Anything, sourceAccountID).Return(mockSourceAccount, nil)
	mockStore.On("GetAccount", mock.Anything, targetAccountID).
		Return(account.Account{}, ErrNoAccountFound)

	// Call the method being tested
	result, err := service.Transfer(context.Background(), mockTransaction)

	// Assert that the expected methods were called on the mock store
	mockStore.AssertExpectations(t)

	// Assert the result and error
	assert.Error(t, err)
	assert.EqualError(t, err, ErrNoAccountFound.Error())

	assert.Equal(t, "", result.ID)
	assert.Equal(t, "", result.SourceAccountID)
	assert.Equal(t, "", result.TargetAccountID)
	assert.Equal(t, 0.0, result.Amount)
	assert.Equal(t, "", result.Currency)
}
