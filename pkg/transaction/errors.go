package transaction

import (
	"errors"
)

// ErrPostingTransaction is used when a transaction could not be created
func ErrPostingTransaction(transactionID string) error {
	return errors.New("could not create a new transaction by ID " + transactionID)
}

// ErrFetchingTransaction is used when a transaction could not be found
func ErrFetchingTransaction(transactionID string) error {
	return errors.New("could not fetch transaction by ID " + transactionID)
}

// ErrUpdateAccount is used when an account could not be updated during a transfer
func ErrUpdateAccount(accountID string) error {
	return errors.New("could not update an account by ID " + accountID)
}

// ErrDeletingTransaction is used when a transaction could not be removed
func ErrDeletingTransaction(transactionID string) error {
	return errors.New("could not delete transaction by ID " + transactionID)
}
