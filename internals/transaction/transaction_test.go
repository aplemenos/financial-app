package transaction

import (
	"context"
	"database/sql"
	"financial-app/pkg/models"
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
) (models.Account, error) {
	args := m.Called(ctx, ID)
	return args.Get(0).(models.Account), args.Error(1)
}

// PostAccount - is a mock implementation of PostAccount method
func (m *MockTransactionStore) PostAccount(
	ctx context.Context, acct models.Account,
) (models.Account, error) {
	args := m.Called(ctx, acct)
	return args.Get(0).(models.Account), args.Error(1)
}

// UpdateAccount - is a mock implementation of UpdateAccount method
func (m *MockTransactionStore) UpdateAccount(
	ctx context.Context, ID string, newAccount models.Account,
) (models.Account, error) {
	args := m.Called(ctx, ID, newAccount)
	return args.Get(0).(models.Account), args.Error(1)
}

// DeleteAccount - is a mock implementation of DeleteAccount method
func (m *MockTransactionStore) DeleteAccount(ctx context.Context, ID string) error {
	args := m.Called(ctx, ID)
	return args.Error(0)
}

// GetTransaction - is a mock implementation of GetTransaction method
func (m *MockTransactionStore) GetTransaction(
	ctx context.Context, ID string,
) (models.Transaction, error) {
	args := m.Called(ctx, ID)
	return args.Get(0).(models.Transaction), args.Error(1)
}

// PostTransaction - is a mock implementation of PostTransaction method
func (m *MockTransactionStore) PostTransaction(
	ctx context.Context, txn models.Transaction,
) (models.Transaction, error) {
	args := m.Called(ctx, txn)
	return args.Get(0).(models.Transaction), args.Error(1)
}

// DeleteTransaction - is a mock implementation of DeleteTransaction method
func (m *MockTransactionStore) DeleteTransaction(ctx context.Context, ID string) error {
	args := m.Called(ctx, ID)
	return args.Error(0)
}

// ExecuteDBTransaction is the mock implementation of ExecuteDBTransaction
func (m *MockTransactionStore) ExecuteDBTransaction(f func(*sql.Tx) error) error {
	// Create a mock transaction
	mockTx := &sql.Tx{}

	// Call the provided function with the mock transaction
	err := f(mockTx)
	if err != nil {
		return err
	}

	// Return a mock success response
	return nil
}

// Ping - is the mock implementation for Ping.
func (m *MockTransactionStore) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestPostTransaction(t *testing.T) {
	// Create a mock transaction store
	mockStore := new(MockTransactionStore)

	// Create the transaction service with the mock store
	service := NewService(mockStore)

	// Set up the mock data for the test case
	sourceAccountID := "1111"
	targetAccountID := "2222"

	// AC 1: Happy path for money transfer between two accounts
	mockSourceAccount := models.Account{
		ID:       sourceAccountID,
		Balance:  500.00,
		Currency: "EUR",
	}

	mockTargetAccount := models.Account{
		ID:       targetAccountID,
		Balance:  200.00,
		Currency: "EUR",
	}

	mockTransaction := models.Transaction{
		SourceAccountID: sourceAccountID,
		TargetAccountID: targetAccountID,
		Amount:          100.00,
		Currency:        "EUR",
	}

	// Set up the expected behavior of the mock store
	mockStore.On("GetAccount", mock.Anything, sourceAccountID).Return(mockSourceAccount, nil)
	mockStore.On("GetAccount", mock.Anything, targetAccountID).Return(mockTargetAccount, nil)
	mockStore.On("UpdateAccount", mock.Anything, sourceAccountID,
		mock.Anything).Return(mockSourceAccount, nil)
	mockStore.On("UpdateAccount", mock.Anything, targetAccountID,
		mock.Anything).Return(mockTargetAccount, nil)
	mockStore.On("PostTransaction", mock.Anything, mockTransaction).Return(mockTransaction, nil)

	// Call the method being tested
	result, err := service.PostTransaction(context.Background(), mockTransaction)

	// Assert the expected behavior
	assert.NoError(t, err)
	assert.Equal(t, mockTransaction, result)
	mockStore.AssertExpectations(t)

	// Test AC 2 - Insufficient balance to process money transfer
	mockTransaction.Amount = 600.00

	_, err = service.PostTransaction(context.Background(), mockTransaction)

	assert.Error(t, err)
	assert.EqualError(t, err, ErrInsufficientBalance.Error())
	mockStore.AssertExpectations(t)

	// AC 4: One or more of the accounts does not exist
	nonExistingAccountID := "3333"
	mockStore.On("GetAccount", mock.Anything, sourceAccountID).Return(mockSourceAccount, nil)
	mockStore.On("GetAccount", mock.Anything,
		nonExistingAccountID).Return(models.Account{}, ErrNoAccountFound)

	mockTransaction.SourceAccountID = sourceAccountID
	mockTransaction.TargetAccountID = nonExistingAccountID

	_, err = service.PostTransaction(context.Background(), mockTransaction)

	assert.Error(t, err)
	assert.EqualError(t, err, ErrNoAccountFound.Error())
	mockStore.AssertExpectations(t)
}
