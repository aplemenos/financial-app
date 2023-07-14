package transaction

import (
	"context"
	"errors"
	"financial-app/domain/account"

	uuid "github.com/satori/go.uuid"
)

// TransactionID uniquely identifies a particular transaction
type TransactionID string

// Transaction is the central class in the domain model
type Transaction struct {
	ID              TransactionID
	SourceAccountID account.AccountID
	TargetAccountID account.AccountID
	Amount          float64
	Currency        account.Currency
}

// NewTransaction creates a new transaction
func NewTransaction(
	id TransactionID, sID, tID account.AccountID, amt float64, cur account.Currency,
) *Transaction {
	return &Transaction{
		ID:              id,
		SourceAccountID: sID,
		TargetAccountID: tID,
		Amount:          amt,
		Currency:        cur,
	}
}

// TransactionRepository provides access a transaction store
type TransactionRepository interface {
	Transfer(
		ctx context.Context, txn *Transaction, sacc *account.Account, tacc *account.Account,
	) (*Transaction, error)
	Find(ctx context.Context, id TransactionID) (*Transaction, error)
	FindAll(ctx context.Context) []*Transaction
	Delete(ctx context.Context, id TransactionID) error
	Ping(ctx context.Context) error
}

// ErrPostingTransaction is used when a transaction could not be created
var ErrPostingTransaction = errors.New("could not create a new transaction")

// ErrFetchingTransaction is used when a transaction could not be found
func ErrFetchingTransaction(id TransactionID) error {
	return errors.New("could not fetch transaction by ID " + string(id))
}

// ErrUpdateAccount is used when an account could not be updated during a transfer
func ErrUpdateAccount(id account.AccountID) error {
	return errors.New("could not update an account by ID " + string(id))
}

// ErrDeletingTransaction is used when a transaction could not be removed
func ErrDeletingTransaction(id TransactionID) error {
	return errors.New("could not delete transaction by ID " + string(id))
}

// NextTransactionID generates a new transaction ID.
func NextTransactionID() TransactionID {
	return TransactionID(uuid.NewV4().String())
}
