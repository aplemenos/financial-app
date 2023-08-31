package account

import (
	"errors"
	"strings"
)

// ErrEmptyAccountList is used when the given account list is empty
var ErrEmptyAccountList = errors.New("account list is empty")

// ErrPostingAccount is used when an account could not be created
func ErrPostingAccount(id string) error {
	return errors.New("could not create a new account by ID " + id)
}

// ErrScanAccount is used when an account could not be scanned
func ErrScanAccounts(ids []string) error {
	return errors.New("could not scan the accounts by IDs " + strings.Join(ids, ","))
}

// ErrQueryingAccounts is used when one of the accounts could not be queried
func ErrQueryingAccounts(ids []string) error {
	return errors.New("could not query the requested accounts by IDs " + strings.Join(ids, ","))
}

// ErrFetchingAccounts is used when one of the accounts could not be fetched
func ErrFetchingAccounts(ids []string) error {
	return errors.New("could not fetch the requested accounts by IDs " + strings.Join(ids, ","))
}

// ErrFetchingAccount is used when an account could not be found
func ErrFetchingAccount(id string) error {
	return errors.New("could not fetch account by ID " + id)
}

// ErrDeletingAccount is used when an account could not be removed
func ErrDeletingAccount(id string) error {
	return errors.New("could not delete account by ID " + id)
}
