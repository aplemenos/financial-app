package transactions

import (
	"context"
	"errors"
	account "financial-app/pkg/accounts"
	"fmt"

	uuid "github.com/satori/go.uuid"
)

// Transaction is a read model for transaction views
type Transaction struct {
	ID              string  `json:"id"`
	SourceAccountID string  `json:"source_account_id"`
	TargetAccountID string  `json:"target_account_id"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
}

// Service is the interface that provides transaction methods
type Service interface {
	// Load returns a read model of a transaction
	Load(ctx context.Context, id string) (Transaction, error)

	// Transfer makes a transaction between two accounts
	Transfer(ctx context.Context, txn Transaction) (Transaction, error)

	// LoadAll returns a list of transactions have been completed
	LoadAll(ctx context.Context) []Transaction

	// Clean deletes a transaction
	Clean(ctx context.Context, id string) error
}

func (s *service) Load(
	ctx context.Context, id string,
) (Transaction, error) {
	// Calls the repository passing in the context
	txn, err := s.transactions.Find(ctx, id)
	if err != nil {
		return Transaction{}, err
	}
	return *txn, nil
}

func (s *service) Transfer(
	ctx context.Context, txn Transaction,
) (Transaction, error) {
	// Get the source and target accounts using one database query
	uuids := []string{txn.SourceAccountID, txn.TargetAccountID}
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

	return *transaction, nil
}

func (s *service) LoadAll(ctx context.Context) []Transaction {
	var transactions []Transaction
	for _, t := range s.transactions.FindAll(ctx) {
		transactions = append(transactions, *t)
	}
	return transactions
}

func (s *service) Clean(ctx context.Context, id string) error {
	if err := s.transactions.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

type service struct {
	accounts     account.AccountRepository
	transactions TransactionRepository
}

// NewService creates a transaction service with necessary dependencies
func NewService(
	accounts account.AccountRepository,
	transactions TransactionRepository,
) Service {
	return &service{
		accounts:     accounts,
		transactions: transactions,
	}
}

// nextTransactionID generates a new transaction ID.
func nextTransactionID() string {
	return uuid.NewV4().String()
}

// errInsufficientBalance is used when a transaction could not be performed
// because of insufficient balance
func errInsufficientBalance(bal, amt float64, id string) error {
	b := fmt.Sprintf("%8.2f", bal)
	a := fmt.Sprintf("%8.2f", amt)
	return errors.New("the source amount is insufficient: " + b + " < " +
		a + " for the account " + string(id))
}
