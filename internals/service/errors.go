package service

import (
	"errors"
	"fmt"
)

var (
	ErrPostingAccount     = errors.New("could not create a new account")
	ErrPostingTransaction = errors.New("could not perform the transfer")
)

func ErrFetchingAccount(id string) error {
	return errors.New("could not fetch account by ID " + id)
}

func ErrNoSourceAccountFound(sID string) error {
	return errors.New("no source account found by ID " + sID)
}

func ErrNoTargetAccountFound(tID string) error {
	return errors.New("no target account found by ID " + tID)
}

func ErrDeletingAccount(id string) error {
	return errors.New("could not delete account by ID " + id)
}

func ErrFetchingTransaction(id string) error {
	return errors.New("could not fetch transaction by ID " + id)
}

func ErrInsufficientBalance(balance, amount float64, id string) error {
	b := fmt.Sprintf("%8.2f", balance)
	a := fmt.Sprintf("%8.2f", amount)
	return errors.New("the source amount is insufficient: " + b + " < " +
		a + " for the account " + id)
}

func ErrNoTransactionFound(id string) error {
	return errors.New("no transaction found by ID " + id)
}

func ErrDeletingTransaction(id string) error {
	return errors.New("could not delete transaction by ID " + id)
}
