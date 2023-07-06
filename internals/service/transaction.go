package service

import (
	"context"
	"errors"
	"financial-app/pkg/account"
	"financial-app/pkg/transaction"

	log "github.com/sirupsen/logrus"
)

var (
	ErrFetchingAccount = errors.New("could not fetch account by ID")
	ErrPostingAccount  = errors.New("could not perform account")
	ErrNoAccountFound  = errors.New("no account found")
	ErrDeletingAccount = errors.New("could not delete account")

	ErrFetchingTransaction = errors.New("could not fetch transaction by ID")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrPostingTransaction  = errors.New("could not perform transaction")
	ErrNoTransactionFound  = errors.New("no transaction found")
	ErrDeletingTransaction = errors.New("could not delete transaction")
)

// TransactionStore - defines the interface we need our transaction storage
// layer to implement
type TransactionStore interface {
	GetAccount(context.Context, string) (account.Account, error)
	PostAccount(context.Context, account.Account) (account.Account, error)
	DeleteAccount(context.Context, string) error

	GetTransaction(context.Context, string) (transaction.Transaction, error)
	DeleteTransaction(context.Context, string) error

	Transfer(ctx context.Context, txn transaction.Transaction, sacc account.Account,
		tacc account.Account) (transaction.Transaction, error)

	Ping(context.Context) error
}

// Service - the struct for our transaction service
type Service struct {
	Store TransactionStore
}

// NewService - returns a new transaction service
func NewService(store TransactionStore) *Service {
	return &Service{
		Store: store,
	}
}

// GetAccount - retrieves an account by ID from the store
func (s *Service) GetAccount(ctx context.Context, ID string) (account.Account, error) {
	// Calls the store passing in the context
	acct, err := s.Store.GetAccount(ctx, ID)
	if err != nil {
		log.Errorf("an error occured fetching the account: %s", err.Error())
		return account.Account{}, ErrFetchingAccount
	}
	return acct, nil
}

// PostAccount - create a new account
func (s *Service) PostAccount(
	ctx context.Context, acct account.Account,
) (account.Account, error) {
	acct, err := s.Store.PostAccount(ctx, acct)
	if err != nil {
		log.Errorf("an error occurred performing the account: %s", err.Error())
		return account.Account{}, ErrPostingAccount
	}
	return acct, nil
}

// DeleteAccount- deletes an account from the store by ID
func (s *Service) DeleteAccount(ctx context.Context, ID string) error {
	if err := s.Store.DeleteAccount(ctx, ID); err != nil {
		log.Errorf("an error occurred deleting the account: %s", err.Error())
		return ErrDeletingAccount
	}
	return nil
}

// GetTransaction - retrieves a transaction by ID from the store
func (s *Service) GetTransaction(
	ctx context.Context, ID string,
) (transaction.Transaction, error) {
	// Calls the store passing in the context
	txn, err := s.Store.GetTransaction(ctx, ID)
	if err != nil {
		log.Errorf("an error occured fetching the transaction: %s", err.Error())
		return transaction.Transaction{}, ErrFetchingTransaction
	}
	return txn, nil
}

// Transfer - performs a new transaction
func (s *Service) Transfer(
	ctx context.Context, txn transaction.Transaction,
) (transaction.Transaction, error) {
	log.Info("Perform a new transaction")

	sourceAccount, err := s.Store.GetAccount(ctx, txn.SourceAccountID)
	if err != nil {
		log.Error("no source account found")
		return transaction.Transaction{}, ErrNoAccountFound
	}

	targetAccount, err := s.Store.GetAccount(ctx, txn.TargetAccountID)
	if err != nil {
		log.Error("no target account found")
		return transaction.Transaction{}, ErrNoAccountFound
	}

	// Check if the source account has sufficient balance
	if sourceAccount.Balance < txn.Amount {
		log.Error("insuffient balance: the source amount ", sourceAccount.Balance, " < ",
			txn.Amount)
		return transaction.Transaction{}, ErrInsufficientBalance
	}

	// Debit the balance from the source account
	sourceAccount.Balance -= txn.Amount

	// Credit the balance to the target account
	targetAccount.Balance += txn.Amount

	// Transfer money securely from source to target account
	return s.Store.Transfer(ctx, txn, sourceAccount, targetAccount)
}

// DeleteTransaction - deletes a transaction from the store by ID
func (s *Service) DeleteTransaction(ctx context.Context, ID string) error {
	if err := s.Store.DeleteTransaction(ctx, ID); err != nil {
		log.Errorf("an error occurred deleting the transaction: %s", err.Error())
		return ErrDeletingTransaction
	}
	return nil
}

// AliveCheck - a function that tests we are functionally alive to serve requests
func (s *Service) AliveCheck(ctx context.Context) error {
	log.Info("checking store aliveness")
	return s.Store.Ping(ctx)
}
