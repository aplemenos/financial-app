package accounts

import "context"

// AccountRepository provides access an account store
type AccountRepository interface {
	Store(ctx context.Context, acct *Account) (*Account, error)
	Find(ctx context.Context, id string) (*Account, error)
	FindByIDs(ctx context.Context, ids []string) (map[string]*Account, error)
	FindAll(ctx context.Context) []*Account
	Delete(ctx context.Context, id string) error
}
