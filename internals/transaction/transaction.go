package transaction

import (
	"context"
	"errors"
	"financial-app/pkg/models"

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
	GetAccount(context.Context, string) (models.Account, error)
	PostAccount(context.Context, models.Account) (models.Account, error)
	UpdateAccount(context.Context, string, models.Account) (models.Account, error)
	DeleteAccount(context.Context, string) error

	GetTransaction(context.Context, string) (models.Transaction, error)
	PostTransaction(context.Context, models.Transaction) (models.Transaction, error)
	DeleteTransaction(context.Context, string) error

	// ExecuteDBTransaction(func(*sql.Tx) error) error

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
func (s *Service) GetAccount(ctx context.Context, ID string) (models.Account, error) {
	// Calls the store passing in the context
	acct, err := s.Store.GetAccount(ctx, ID)
	if err != nil {
		log.Errorf("an error occured fetching the account: %s", err.Error())
		return models.Account{}, ErrFetchingAccount
	}
	return acct, nil
}

// PostAccount - create a new account
func (s *Service) PostAccount(
	ctx context.Context, acct models.Account,
) (models.Account, error) {
	acct, err := s.Store.PostAccount(ctx, acct)
	if err != nil {
		log.Errorf("an error occurred performing the account: %s", err.Error())
		return models.Account{}, ErrPostingAccount
	}
	return acct, nil
}

// UpdateAccount - updates a account by ID with new account info
func (s *Service) UpdateAccount(
	ctx context.Context, ID string, newAccount models.Account,
) (models.Account, error) {
	acct, err := s.Store.UpdateAccount(ctx, ID, newAccount)
	if err != nil {
		log.Errorf("an error occurred updating the account: %s", err.Error())
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
) (models.Transaction, error) {
	// Calls the store passing in the context
	txn, err := s.Store.GetTransaction(ctx, ID)
	if err != nil {
		log.Errorf("an error occured fetching the transaction: %s", err.Error())
		return models.Transaction{}, ErrFetchingTransaction
	}
	return txn, nil
}

// PostTransanction - performs a new transaction
func (s *Service) PostTransaction(
	ctx context.Context, txn models.Transaction,
) (models.Transaction, error) {
	log.Info("Perform a new transaction")

	sourceAccount, err := s.Store.GetAccount(ctx, txn.SourceAccountID)
	if err != nil {
		log.Error("no source account found")
		return models.Transaction{}, ErrNoAccountFound
	}

	targetAccount, err := s.Store.GetAccount(ctx, txn.TargetAccountID)
	if err != nil {
		log.Error("no target account found")
		return models.Transaction{}, ErrNoAccountFound
	}

	// Check if the source account has sufficient balance
	if sourceAccount.Balance < txn.Amount {
		log.Error("insuffient balance: the source amount ", sourceAccount.Balance, " < ",
			txn.Amount)
		return models.Transaction{}, ErrInsufficientBalance
	}

	// Debit the balance from the source account
	sourceAccount.Balance -= txn.Amount

	// Credit the balance to the target account
	targetAccount.Balance += txn.Amount

	// Update the account balances in the account store
	_, err = s.Store.UpdateAccount(ctx, txn.SourceAccountID, sourceAccount)
	if err != nil {
		log.Errorf("an error occurred updating the source account: %s", err.Error())
		return models.Transaction{}, err
	}

	_, err = s.Store.UpdateAccount(ctx, txn.TargetAccountID, targetAccount)
	if err != nil {
		log.Errorf("an error occurred updating the target account: %s", err.Error())
		return models.Transaction{}, err
	}

	// Set a new entry of the completed transaction in the transaction store
	txn, err = s.Store.PostTransaction(ctx, txn)
	if err != nil {
		log.Errorf("an error occurred performing the transaction: %s", err.Error())
		return models.Transaction{}, ErrPostingTransaction
	}
	return txn, nil
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
