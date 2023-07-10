package service

import (
	"context"
	"financial-app/pkg/account"
	"financial-app/pkg/transaction"
	"financial-app/util/logger"
	"financial-app/util/multiplelock"
)

// TransactionStore - defines the interface we need our transaction storage
// layer to implement
type TransactionStore interface {
	GetAccount(context.Context, string) (account.Account, error)
	GetAccounts(context.Context, []string) (map[string]account.Account, error)
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
	MLock multiplelock.MultipleLock
	Store TransactionStore
}

// NewService - returns a new transaction service
func NewService(store TransactionStore) *Service {
	return &Service{
		// Thread safe - it is only create one instance
		MLock: multiplelock.NewMultipleLock(),
		Store: store,
	}
}

// GetAccount - retrieves an account by ID from the store
func (s *Service) GetAccount(ctx context.Context, id string) (account.Account, error) {
	log := logger.NewLoggerFromReqIDCtx(ctx, nil)

	// Calls the store passing in the context
	acct, err := s.Store.GetAccount(ctx, id)
	if err != nil {
		log.Errorf("%s", err.Error())
		return account.Account{}, ErrFetchingAccount(id)
	}
	return acct, nil
}

// PostAccount - create a new account
func (s *Service) PostAccount(
	ctx context.Context, acct account.Account,
) (account.Account, error) {
	log := logger.NewLoggerFromReqIDCtx(ctx, nil)

	acct, err := s.Store.PostAccount(ctx, acct)
	if err != nil {
		log.Errorf("an error occurred performing the account: %s", err.Error())
		return account.Account{}, ErrPostingAccount
	}
	return acct, nil
}

// DeleteAccount- deletes an account from the store by ID
func (s *Service) DeleteAccount(ctx context.Context, id string) error {
	log := logger.NewLoggerFromReqIDCtx(ctx, nil)

	if err := s.Store.DeleteAccount(ctx, id); err != nil {
		log.Errorf("an error occurred deleting the account: %s", err.Error())
		return ErrDeletingAccount(id)
	}
	return nil
}

// GetTransaction - retrieves a transaction by ID from the store
func (s *Service) GetTransaction(
	ctx context.Context, id string,
) (transaction.Transaction, error) {
	log := logger.NewLoggerFromReqIDCtx(ctx, nil)

	// Calls the store passing in the context
	txn, err := s.Store.GetTransaction(ctx, id)
	if err != nil {
		log.Errorf("%s", err.Error())
		return transaction.Transaction{}, ErrFetchingTransaction(id)
	}
	return txn, nil
}

// Transfer - performs a new transaction
func (s *Service) Transfer(
	ctx context.Context, txn transaction.Transaction,
) (transaction.Transaction, error) {
	log := logger.NewLoggerFromReqIDCtx(ctx, nil)
	log.Info("Perform a secure tranfer")

	s.MLock.Lock(txn.SourceAccountID)
	s.MLock.Unlock(txn.SourceAccountID)

	// Get the source and target accounts using one database query
	uuids := []string{txn.SourceAccountID, txn.TargetAccountID}
	accounts, err := s.Store.GetAccounts(ctx, uuids)
	if err != nil {
		log.Errorf("no account found: %s", err.Error())
		return transaction.Transaction{}, ErrNoAccountsFound(uuids)
	}

	sourceAccount := accounts[txn.SourceAccountID]
	targetAccount := accounts[txn.TargetAccountID]

	// Check if the source account has sufficient balance
	if sourceAccount.Balance < txn.Amount {
		log.Error("insuffient balance: the source amount is ", sourceAccount.Balance, " < ",
			txn.Amount, " for the account with id ", sourceAccount.ID)
		return transaction.Transaction{},
			ErrInsufficientBalance(sourceAccount.Balance, txn.Amount, sourceAccount.ID)
	}

	// Debit the balance from the source account
	sourceAccount.Balance -= txn.Amount

	// Credit the balance to the target account
	targetAccount.Balance += txn.Amount

	// Transfer money securely from source to target account
	txn, err = s.Store.Transfer(ctx, txn, sourceAccount, targetAccount)
	if err != nil {
		log.Errorf("could not perform transfer: %s", err.Error())
		return transaction.Transaction{}, ErrPostingTransaction
	}

	return txn, nil
}

// DeleteTransaction - deletes a transaction from the store by ID
func (s *Service) DeleteTransaction(ctx context.Context, id string) error {
	log := logger.NewLoggerFromReqIDCtx(ctx, nil)

	if err := s.Store.DeleteTransaction(ctx, id); err != nil {
		log.Errorf("an error occurred deleting the transaction: %s", err.Error())
		return ErrDeletingTransaction(id)
	}
	return nil
}

// AliveCheck - a function that tests we are functionally alive to serve requests
func (s *Service) AliveCheck(ctx context.Context) error {
	log := logger.NewLoggerFromReqIDCtx(ctx, nil)
	log.Info("checking store aliveness")

	return s.Store.Ping(ctx)
}
