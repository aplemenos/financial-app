package account

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockAccountRepository is a mock implementation of AccountRepository
type mockAccountRepository struct {
	Accounts map[AccountID]*Account
}

func (m *mockAccountRepository) Store(
	ctx context.Context, acct *Account,
) (*Account, error) {
	m.Accounts[acct.ID] = acct
	return acct, nil
}

func (m *mockAccountRepository) Find(
	ctx context.Context, id AccountID,
) (*Account, error) {
	if acct, ok := m.Accounts[id]; ok {
		return acct, nil
	}
	return nil, ErrFetchingAccount(id)
}

func (m *mockAccountRepository) FindByIDs(
	ctx context.Context, ids []AccountID,
) (map[AccountID]*Account, error) {
	accounts := make(map[AccountID]*Account)
	for _, id := range ids {
		if acct, ok := m.Accounts[id]; ok {
			accounts[id] = acct
		}
	}

	return accounts, nil
}

func (m *mockAccountRepository) FindAll(
	ctx context.Context,
) ([]*Account, error) {
	accounts := make([]*Account, 0, len(m.Accounts))
	for _, txn := range m.Accounts {
		accounts = append(accounts, txn)
	}
	return accounts, nil
}

func (m *mockAccountRepository) Delete(
	ctx context.Context, id AccountID,
) error {
	if _, ok := m.Accounts[id]; ok {
		delete(m.Accounts, id)
		return nil
	}
	return ErrDeletingAccount(id)
}

func TestAccount_New(t *testing.T) {
	id := AccountID("1111")
	balance := 100.0
	cur := Currency(EUR)

	acct := NewAccount(id, balance, cur)

	assert.Equal(t, id, acct.ID, "Account ID should match")
	assert.Equal(t, balance, acct.Balance, "Account balance should match")
	assert.Equal(t, cur, acct.Currency, "Account currency should match")
}

func TestAccount_IsSupportedCurrency(t *testing.T) {
	acctUSD := &Account{Currency: USD}
	acctEUR := &Account{Currency: EUR}
	acctInvalid := &Account{Currency: "GBP"}

	assert.True(t, IsSupportedCurrency(acctUSD.Currency), "USD currency should be supported")
	assert.True(t, IsSupportedCurrency(acctEUR.Currency), "EUR currency should be supported")
	assert.False(t, IsSupportedCurrency(acctInvalid.Currency), "GBP currency should not be supported")
}

func TestAccountRepository_Store(t *testing.T) {
	repo := &mockAccountRepository{Accounts: make(map[AccountID]*Account)}

	id := NextAccountID()
	balance := 100.0
	cur := Currency(EUR)
	acct := NewAccount(id, balance, cur)

	storedAcct, err := repo.Store(context.Background(), acct)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, acct, storedAcct, "Stored account should match the provided account")
}

func TestAccountRepository_Find(t *testing.T) {
	repo := &mockAccountRepository{
		Accounts: map[AccountID]*Account{
			AccountID("1111"): {ID: AccountID("1111")},
		},
	}

	id := AccountID("1111")
	acct, err := repo.Find(context.Background(), id)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, id, acct.ID, "Found account ID should match the provided ID")
}

func TestAccountRepository_FindByIDs(t *testing.T) {
	repo := &mockAccountRepository{
		Accounts: map[AccountID]*Account{
			AccountID("1111"): {ID: AccountID("1111")},
			AccountID("2222"): {ID: AccountID("2222")},
		},
	}

	ids := []AccountID{AccountID("1111"), AccountID("2222")}
	accounts, err := repo.FindByIDs(context.Background(), ids)

	assert.NoError(t, err, "Error should be nil")
	assert.Equal(t, 2, len(accounts), "Number of found accounts should be 2")
	assert.Equal(
		t,
		AccountID("1111"),
		accounts["1111"].ID,
		"Found account ID should match the provided ID",
	)
}

func TestAccountRepository_Delete(t *testing.T) {
	repo := &mockAccountRepository{
		Accounts: map[AccountID]*Account{
			AccountID("1111"): {ID: AccountID("1111")},
		},
	}

	id := AccountID("1111")
	err := repo.Delete(context.Background(), id)

	assert.NoError(t, err, "Error should be nil")
	_, exists := repo.Accounts[id]
	assert.False(t, exists, "Account should be deleted")
}

func TestAccountRepository_FindAll(t *testing.T) {
	repo := &mockAccountRepository{
		Accounts: map[AccountID]*Account{
			AccountID("1111"): {ID: AccountID("1111")},
			AccountID("2222"): {ID: AccountID("2222")},
			AccountID("3333"): {ID: AccountID("3333")},
		},
	}

	accounts, err := repo.FindAll(context.Background())

	assert.NoError(t, err, "Error should be nil")
	assert.Len(t, accounts, 3, "Number of returned accounts should be 3")
	assert.Contains(
		t,
		accounts,
		&Account{ID: AccountID("1111")},
		"Account 1 should be present",
	)
	assert.Contains(
		t,
		accounts,
		&Account{ID: AccountID("2222")},
		"Account 2 should be present",
	)
	assert.Contains(
		t,
		accounts,
		&Account{ID: AccountID("3333")},
		"Account 3 should be present",
	)
}

func TestAccountRepository_FindAllAccountsEmpty(t *testing.T) {
	repo := &mockAccountRepository{
		Accounts: make(map[AccountID]*Account),
	}

	accounts, err := repo.FindAll(context.Background())

	assert.NoError(t, err, "Error should be nil")
	assert.Empty(t, accounts, "Returned accounts should be empty")
}
