package service

import (
	"errors"
	"financial-app/internals/transaction/repo/kvs"
	"financial-app/pkg/models"
	"log"
)

type Transaction struct {
	Accounts     *kvs.AccountStore
	Transactions *kvs.TransactionStore
}

// NewTransaction creates a new instance of Transaction
func NewTransaction(accountStore *kvs.AccountStore) *Transaction {
	return &Transaction{
		Accounts: accountStore,
	}
}

func (t *Transaction) CreateTransanction(transaction models.Transaction) error {
	log.Println("Validate the transaction")
	// Validate the transaction
	err := t.validateTransaction(transaction)
	if err != nil {
		return err
	}
	log.Println("Transaction validation passed")

	log.Println("Process the transaction")
	// Process the transaction
	err = t.processTransaction(transaction)
	if err != nil {
		return err
	}

	return nil
}

// ValidateTransaction performs validation checks based on the acceptance criteria
// Return an error if any validation fails
func (t *Transaction) validateTransaction(transaction models.Transaction) error {

	// Check if the source account exists
	sourceAccount, err := t.Accounts.Get(transaction.SourceAccountID)
	if err != nil {
		return errors.New("source account does not exist")
	}

	// Check if the target account exists
	targetAccount, err := t.Accounts.Get(transaction.TargetAccountID)
	if err != nil {
		return errors.New("target account does not exist")
	}

	// Validate the source account is not the same with the target one
	if sourceAccount.ID == targetAccount.ID {
		return errors.New("transfer between same account")
	}

	// Validate the transaction amount
	if transaction.Amount < 0 {
		return errors.New("transaction amount must be positive")
	}

	// Check if the source account has sufficient balance
	if sourceAccount.Balance < transaction.Amount {
		return errors.New("insufficient balance in the source account")
	}

	return nil
}

// ProcessTransaction performs the actual transaction between the accounts
// Update the account balances accordingly
func (t *Transaction) processTransaction(transaction models.Transaction) error {
	sourceAccount, _ := t.Accounts.Get(transaction.SourceAccountID)
	targetAccount, _ := t.Accounts.Get(transaction.SourceAccountID)

	// Debit the balance from the source account
	sourceAccount.Balance -= transaction.Amount

	// Credit the balance to the target account
	targetAccount.Balance += transaction.Amount

	// Update the account balances in the account store
	t.Accounts.Set(transaction.SourceAccountID, *sourceAccount)
	t.Accounts.Set(transaction.TargetAccountID, *targetAccount)
	// Set a new entry of the completed transaction in the transaction store
	t.Transactions.Set(transaction.ID, transaction)

	return nil
}

func (t *Transaction) GetTransanction(transactionID string) (*models.Transaction, error) {
	log.Println("Get the transaction ", transactionID)

	transaction, err := t.Transactions.Get(transactionID)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}
