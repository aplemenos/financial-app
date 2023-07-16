package register

import (
	"context"
	"financial-app/domain/account"
	"testing"
	"time"

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

func TestService_LoadAccount(t *testing.T) {
	accountID := account.AccountID("1111")

	mockAccount := &account.Account{
		ID:        accountID,
		Balance:   1000.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
	}

	mockAccounts := map[account.AccountID]*account.Account{
		accountID: mockAccount,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccounts,
	}

	service := NewService(mockAccountRepository)

	loadedAccount, err := service.LoadAccount(context.Background(), accountID)

	expectedAccount := assemble(mockAccount)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, expectedAccount, loadedAccount, "Loaded account should match the mock account")
}

func TestService_Store(t *testing.T) {
	accountID := account.AccountID("1111")

	mockAccount := &account.Account{
		ID:        accountID,
		Balance:   1000.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
	}

	mockAccounts := map[account.AccountID]*account.Account{
		accountID: mockAccount,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccounts,
	}

	service := NewService(mockAccountRepository)

	newAccount, err := service.Register(context.Background(), *mockAccount)

	expectedAccount := assemble(mockAccount)

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
		mockAccount,
		mockAccountRepository.Accounts[accountID],
		"Stored account should match the mock account",
	)
}

func TestService_Accounts(t *testing.T) {
	accountID1 := account.AccountID("1111")
	accountID2 := account.AccountID("2222")

	mockAccount1 := &account.Account{
		ID:        accountID1,
		Balance:   1000.0,
		Currency:  "USD",
		CreatedAt: time.Now(),
	}

	mockAccount2 := &account.Account{
		ID:        accountID2,
		Balance:   2000.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}

	mockAccounts := map[account.AccountID]*account.Account{
		accountID1: mockAccount1,
		accountID2: mockAccount2,
	}

	mockAccountRepository := &mockAccountRepository{
		Accounts: mockAccounts,
	}

	service := NewService(mockAccountRepository)

	accounts := service.Accounts(context.Background())

	assert.Len(t, accounts, 2, "Number of returned accounts should be 2")
	assert.Contains(t, accounts, assemble(mockAccount1), "Account 1 should be present")
	assert.Contains(t, accounts, assemble(mockAccount2), "Account 2 should be present")
}

func TestService_Clean(t *testing.T) {
	mockAccountID := account.AccountID("1111")

	mockAccountRepository := &mockAccountRepository{
		Accounts: map[account.AccountID]*account.Account{
			mockAccountID: {},
		},
	}

	service := NewService(mockAccountRepository)

	err := service.Clean(context.Background(), mockAccountID)

	assert.NoError(t, err, "Error should be nil")
	_, exists := mockAccountRepository.Accounts[mockAccountID]
	assert.False(t, exists, "Account should be deleted")
}
