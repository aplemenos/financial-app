package transfer

import (
	"context"
	"errors"
	"financial-app/domain/account"
	"financial-app/domain/transaction"
	"financial-app/multiplelock"
	"fmt"
)

// Service is the interface that provides transfer methods
type Service interface {
	// LoadTransaction returns a read model of a transaction
	LoadTransaction(ctx context.Context, id transaction.TransactionID) (Transaction, error)

	// Transfer makes a transaction between two accounts
	Transfer(ctx context.Context, txn transaction.Transaction) (Transaction, error)

	// Transactions returns a list of transactions have been completed
	Transactions(ctx context.Context) []Transaction

	// Clean deletes a transaction
	Clean(ctx context.Context, id transaction.TransactionID) error

	// Check repository aliveness
	Alive(ctx context.Context) error
}

func (s *service) LoadTransaction(
	ctx context.Context, id transaction.TransactionID,
) (Transaction, error) {
	// Calls the repository passing in the context
	txn, err := s.transactions.Find(ctx, id)
	if err != nil {
		return Transaction{}, err
	}
	return assemble(txn), nil
}

func (s *service) Transfer(
	ctx context.Context, txn transaction.Transaction,
) (Transaction, error) {
	s.mlock.Lock(txn.SourceAccountID)
	s.mlock.Unlock(txn.SourceAccountID)

	// Get the source and target accounts using one database query
	uuids := []account.AccountID{txn.SourceAccountID, txn.TargetAccountID}
	accounts, err := s.accounts.FindByIDs(ctx, uuids)
	if err != nil {
		return Transaction{}, err
	}

	sourceAccount := accounts[txn.SourceAccountID]
	targetAccount := accounts[txn.TargetAccountID]

	if sourceAccount == nil {
		return Transaction{}, account.ErrFetchingAccount(txn.SourceAccountID)
	}

	if targetAccount == nil {
		return Transaction{}, account.ErrFetchingAccount(txn.TargetAccountID)
	}

	// Check if the source account has sufficient balance
	if sourceAccount.Balance < txn.Amount {
		return Transaction{},
			errInsufficientBalance(sourceAccount.Balance, txn.Amount, txn.SourceAccountID)
	}

	// Debit the balance from the source account
	sourceAccount.Balance -= txn.Amount

	// Credit the balance to the target account
	targetAccount.Balance += txn.Amount

	// Transfer money from source to target account
	transaction, err := s.transactions.Transfer(ctx, &txn, sourceAccount, targetAccount)
	if err != nil {
		return Transaction{}, err
	}

	return assemble(transaction), nil
}

func (s *service) Transactions(ctx context.Context) []Transaction {
	var result []Transaction
	for _, t := range s.transactions.FindAll(ctx) {
		result = append(result, assemble(t))
	}
	return result
}

func (s *service) Clean(ctx context.Context, id transaction.TransactionID) error {
	if err := s.transactions.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *service) Alive(ctx context.Context) error {
	if err := s.transactions.Ping(ctx); err != nil {
		return err
	}
	return nil
}

type service struct {
	accounts     account.AccountRepository
	transactions transaction.TransactionRepository
	mlock        multiplelock.MultipleLock
}

// NewService creates a tranfer service with necessary dependencies
func NewService(
	accounts account.AccountRepository,
	transactions transaction.TransactionRepository,
	mlock multiplelock.MultipleLock,
) Service {
	return &service{
		accounts:     accounts,
		transactions: transactions,
		mlock:        mlock,
	}
}

// Transaction is a read model for transaction views
type Transaction struct {
	ID              string  `json:"id"`
	SourceAccountID string  `json:"source_account_id"`
	TargetAccountID string  `json:"target_account_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
}

func assemble(t *transaction.Transaction) Transaction {
	return Transaction{
		ID:              string(t.ID),
		SourceAccountID: string(t.SourceAccountID),
		TargetAccountID: string(t.TargetAccountID),
		Amount:          t.Amount,
		Currency:        string(t.Currency),
	}
}

// errInsufficientBalance is used when a transaction could not be performed
// because of insufficient balance
func errInsufficientBalance(bal, amt float64, id account.AccountID) error {
	b := fmt.Sprintf("%8.2f", bal)
	a := fmt.Sprintf("%8.2f", amt)
	return errors.New("the source amount is insufficient: " + b + " < " +
		a + " for the account " + string(id))
}
