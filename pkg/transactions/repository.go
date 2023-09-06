package transactions

import (
	"context"
	"financial-app/pkg/accounts"
)

// TransactionRepository provides access a transaction store
type TransactionRepository interface {
	Transfer(
		ctx context.Context, txn *Transaction, sacc *accounts.Account, tacc *accounts.Account,
	) (*Transaction, error)
	Find(ctx context.Context, id string) (*Transaction, error)
	FindAll(ctx context.Context) []*Transaction
	Delete(ctx context.Context, id string) error
}
