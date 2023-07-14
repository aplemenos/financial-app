package account

import (
	"context"
	"errors"
	"time"

	uuid "github.com/satori/go.uuid"
)

// Constants for all supported currencies
const (
	USD = "USD"
	EUR = "EUR"
)

// AccountID uniquely identifies a particular account
type AccountID string

// Currency is the supported medium of exchange
type Currency string

// Account is the central class in the domain model
type Account struct {
	ID        AccountID
	Balance   float64
	Currency  Currency
	CreatedAt time.Time
}

// NewAccount creates a new account
func NewAccount(id AccountID, bal float64, cur Currency) *Account {
	return &Account{
		ID:       id,
		Balance:  bal,
		Currency: cur,
	}
}

// AccountRepository provides access an account store
type AccountRepository interface {
	Store(ctx context.Context, acct *Account) (*Account, error)
	Find(ctx context.Context, id AccountID) (*Account, error)
	FindByIDs(ctx context.Context, ids []AccountID) (map[AccountID]*Account, error)
	FindAll(ctx context.Context) []*Account
	Delete(ctx context.Context, id AccountID) error
}

// ErrPostingAccount is used when an account could not be created
var ErrPostingAccount = errors.New("could not create a new account")

// ErrEmptyAccountList is used when the given account list is empty
var ErrEmptyAccountList = errors.New("account list is empty")

// ErrScanAccount is used when an account could not be scanned
var ErrScanAccount = errors.New("could not scan an account")

// ErrQueryingAccounts is used when one of the accounts could not be queried
var ErrQueryingAccounts = errors.New("could not query the requested accounts")

// ErrFetchingAccounts is used when one of the accounts could not be fetched
var ErrFetchingAccounts = errors.New("could not fetch the requested accounts")

// ErrFetchingAccount is used when an account could not be found
func ErrFetchingAccount(id AccountID) error {
	return errors.New("could not fetch account by ID " + string(id))
}

// ErrDeletingAccount is used when an account could not be removed
func ErrDeletingAccount(id AccountID) error {
	return errors.New("could not delete account by ID " + string(id))
}

// NextAccountID generates a new account ID.
func NextAccountID() AccountID {
	return AccountID(uuid.NewV4().String())
}

// IsSupportedCurrency returns true if the currency is supported
func IsSupportedCurrency(c Currency) bool {
	switch c {
	case USD, EUR:
		return true
	}
	return false
}
