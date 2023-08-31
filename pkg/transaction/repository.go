package transaction

import (
	"context"
	"financial-app/pkg/account"
)

// TransactionRepository provides access a transaction store
type TransactionRepository interface {
	Transfer(
		ctx context.Context, txn *Transaction, sacc *account.Account, tacc *account.Account,
	) (*Transaction, error)
	Find(ctx context.Context, id string) (*Transaction, error)
	FindAll(ctx context.Context) []*Transaction
	Delete(ctx context.Context, id string) error
}
