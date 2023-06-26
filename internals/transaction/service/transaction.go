package service

import (
	"errors"
	"financial-app/internals/transaction/repo/kvs"
	"financial-app/pkg/models"
	"log"

	"github.com/google/uuid"
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

func (t *Transaction) CreateTransanction(transaction models.Transaction) (string, error) {
	log.Println("Validate the transaction")
	// Validate the transaction
	err := t.validateTransaction(transaction)
	if err != nil {
		return "", err
	}
	log.Println("Transaction validation passed")

	log.Println("Process the transaction")
	// Process the transaction
	transactionID := t.processTransaction(transaction)
	log.Println("Transaction completed")

	return transactionID, nil
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
// Insert a new transaction in store
func (t *Transaction) processTransaction(transaction models.Transaction) string {
	sourceAccount, _ := t.Accounts.Get(transaction.SourceAccountID)
	targetAccount, _ := t.Accounts.Get(transaction.TargetAccountID)

	// Debit the balance from the source account
	sourceAccount.Balance -= transaction.Amount

	// Credit the balance to the target account
	targetAccount.Balance += transaction.Amount

	// Update the account balances in the account store
	t.Accounts.Set(transaction.SourceAccountID, *sourceAccount)
	t.Accounts.Set(transaction.TargetAccountID, *targetAccount)
	// Set a new entry of the completed transaction in the transaction store
	uuid := uuid.New() // Generate a new UUIDv4
	transaction.ID = uuid.String()
	t.Transactions.Set(transaction.ID, transaction)

	return transaction.ID
}

func (t *Transaction) GetTransanction(transactionID string) (*models.Transaction, error) {
	log.Println("Get the transaction ", transactionID)

	transaction, err := t.Transactions.Get(transactionID)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (t *Transaction) GetTransactions() []models.Transaction {
	log.Println("Get all transactions")

	return t.Transactions.GetAll()
}
