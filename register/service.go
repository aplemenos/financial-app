package register

import (
	"context"
	"financial-app/domain/account"
	"time"
)

// Service is the interface that provides account methods
type Service interface {
	// LoadAccount returns a read model of an account
	LoadAccount(ctx context.Context, id account.AccountID) (Account, error)

	// Register registers a new account
	Register(ctx context.Context, acct account.Account) (Account, error)

	// Accounts returns a list of accounts have been registered
	Accounts(ctx context.Context) []Account

	// Clean deletes an account
	Clean(ctx context.Context, id account.AccountID) error
}

func (s *service) LoadAccount(
	ctx context.Context, id account.AccountID,
) (Account, error) {
	// Calls the repository passing in the context
	acct, err := s.accounts.Find(ctx, id)
	if err != nil {
		return Account{}, err
	}
	return assemble(acct), nil
}

func (s *service) Register(
	ctx context.Context, acct account.Account,
) (Account, error) {
	// Store the new account to the repository
	account, err := s.accounts.Store(ctx, &acct)
	if err != nil {
		return Account{}, err
	}
	return assemble(account), nil
}

func (s *service) Accounts(ctx context.Context) []Account {
	var result []Account
	for _, a := range s.accounts.FindAll(ctx) {
		result = append(result, assemble(a))
	}
	return result
}

func (s *service) Clean(ctx context.Context, id account.AccountID) error {
	if err := s.accounts.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

type service struct {
	accounts account.AccountRepository
}

// NewService creates an account service with necessary dependencies
func NewService(
	accounts account.AccountRepository,
) Service {
	return &service{
		accounts: accounts,
	}
}

// Account is a read model for account views
type Account struct {
	ID        string    `json:"id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

func assemble(a *account.Account) Account {
	return Account{
		ID:        string(a.ID),
		Balance:   a.Balance,
		Currency:  string(a.Currency),
		CreatedAt: a.CreatedAt,
	}
}
