package account

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockAccountRepository struct {
	Accounts map[string]*Account
}

func (m *mockAccountRepository) Find(
	ctx context.Context, id string,
) (*Account, error) {
	if acct, ok := m.Accounts[id]; ok {
		return acct, nil
	}
	return nil, ErrFetchingAccount(id)
}

func (m *mockAccountRepository) FindByIDs(
	ctx context.Context, ids []string,
) (map[string]*Account, error) {
	accounts := make(map[string]*Account)
	for _, id := range ids {
		if acct, ok := m.Accounts[id]; ok {
			accounts[id] = acct
		}
	}
	return accounts, nil
}

func (m *mockAccountRepository) Store(
	ctx context.Context, acct *Account,
) (*Account, error) {
	m.Accounts[acct.ID] = acct
	return acct, nil
}

func (m *mockAccountRepository) FindAll(
	ctx context.Context,
) []*Account {
	accounts := make([]*Account, 0, len(m.Accounts))
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
	return ErrDeletingAccount(id)
}

func TestService_LoadAccount(t *testing.T) {
	accountID := "1111"

	expectedAccount := Account{
		ID:        accountID,
		Balance:   1000.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
	}

	mockAccount := map[string]*Account{
		accountID: &expectedAccount,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccount,
	}

	service := NewService(mockAccountRepository)

	loadedAccount, err := service.LoadAccount(context.Background(), accountID)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, expectedAccount, loadedAccount, "Loaded account should match the mock account")
}

func TestService_Store(t *testing.T) {
	accountID := "1111"

	expectedAccount := Account{
		ID:        accountID,
		Balance:   1000.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
	}

	mockAccounts := map[string]*Account{
		accountID: &expectedAccount,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccounts,
	}

	service := NewService(mockAccountRepository)

	newAccount, err := service.Register(context.Background(), expectedAccount)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, expectedAccount, newAccount, "Stored account should match the mock account")

	assert.Contains(
		t,
		mockAccountRepository.Accounts,
		accountID,
		"Account should be stored in the repository",
	)
	assert.Equal(
		t,
		expectedAccount,
		*mockAccountRepository.Accounts[accountID],
		"Stored account should match the mock account",
	)
}

func TestService_Accounts(t *testing.T) {
	accountID1 := "1111"
	accountID2 := "2222"

	expectedAccount1 := Account{
		ID:        accountID1,
		Balance:   1000.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
	}

	expectedAccount2 := Account{
		ID:        accountID2,
		Balance:   2000.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}

	mockAccounts := map[string]*Account{
		accountID1: &expectedAccount1,
		accountID2: &expectedAccount2,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccounts,
	}

	service := NewService(mockAccountRepository)

	accounts := service.Accounts(context.Background())

	assert.Len(t, accounts, 2, "Number of returned accounts should be 2")
	assert.Contains(t, accounts, expectedAccount1, "Account 1 should be present")
	assert.Contains(t, accounts, expectedAccount2, "Account 2 should be present")
}

func TestService_Clean(t *testing.T) {
	mockAccountID := "1111"

	mockAccountRepository := &mockAccountRepository{
		Accounts: map[string]*Account{
			mockAccountID: {},
		},
	}

	service := NewService(mockAccountRepository)

	err := service.Clean(context.Background(), mockAccountID)

	assert.NoError(t, err, "Error should be nil")
	_, exists := mockAccountRepository.Accounts[mockAccountID]
	assert.False(t, exists, "Account should be deleted")
}
