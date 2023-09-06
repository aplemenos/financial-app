package accounts

import (
	"context"
	"time"

	uuid "github.com/satori/go.uuid"
)

// Constants for all supported currencies
const (
	USD = "USD"
	EUR = "EUR"
)

// Account is a read model for account views
type Account struct {
	ID        string    `json:"id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

// Service is the interface that provides account methods
type Service interface {
	// LoadAccount returns a read model of an account
	LoadAccount(ctx context.Context, id string) (Account, error)

	// Register registers a new account
	Register(ctx context.Context, acct Account) (Account, error)

	// Accounts returns a list of accounts have been registered
	Accounts(ctx context.Context) []Account

	// Clean deletes an account
	Clean(ctx context.Context, id string) error
}

func (s *service) LoadAccount(
	ctx context.Context, id string,
) (Account, error) {
	// Calls the repository passing in the context
	account, err := s.accounts.Find(ctx, id)
	if err != nil {
		return Account{}, err
	}
	return *account, nil
}

func (s *service) Register(
	ctx context.Context, acct Account,
) (Account, error) {
	// Store the new account to the repository
	account, err := s.accounts.Store(ctx, &acct)
	if err != nil {
		return Account{}, err
	}
	return *account, nil
}

func (s *service) Accounts(ctx context.Context) []Account {
	var accounts []Account
	for _, a := range s.accounts.FindAll(ctx) {
		accounts = append(accounts, *a)
	}
	return accounts
}

func (s *service) Clean(ctx context.Context, id string) error {
	if err := s.accounts.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

type service struct {
	accounts AccountRepository
}

// NewService creates an account service with necessary dependencies
func NewService(
	accounts AccountRepository,
) Service {
	return &service{
		accounts: accounts,
	}
}

// nextAccountID generates a new account ID.
func nextAccountID() string {
	return uuid.NewV4().String()
}

// IsSupportedCurrency returns true if the currency is supported
func IsSupportedCurrency(c string) bool {
	switch c {
	case USD, EUR:
		return true
	}
	return false
}
