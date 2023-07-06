package database

import (
	"context"
	"database/sql"
	"financial-app/pkg/account"
	"financial-app/pkg/transaction"
	"financial-app/util"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetAccount(t *testing.T) {
	// Create a new SQL mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	// Create a new Database instance with the mock DB connection
	d := Database{Client: db}

	// Define the expected account and row data
	expectedID := "1111"
	expectedBalance := 100.0
	expectedCurrency := util.EUR
	expectedCreatedAt := sql.NullTime{Valid: true}

	// Add the expected SQL query and result to the mock
	mock.ExpectQuery("SELECT id, balance, currency, created_at FROM accounts WHERE id =").
		WithArgs(expectedID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "balance", "currency", "created_at"}).
			AddRow(expectedID, expectedBalance, expectedCurrency, expectedCreatedAt))

	// Call the method being tested
	result, err := d.GetAccount(context.Background(), expectedID)

	// Assert the expected result
	assert.NoError(t, err)
	expectedAccount := account.Account{
		ID:        expectedID,
		Balance:   expectedBalance,
		Currency:  expectedCurrency,
		CreatedAt: expectedCreatedAt.Time,
	}
	assert.Equal(t, expectedAccount, result)

	// Assert that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostAccount(t *testing.T) {
	// Create a new SQL mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	// Create a new Database instance with the mock DB connection
	d := Database{Client: db}

	// Define the expected account and row data
	expectedID := "1111"
	expectedBalance := 100.0
	expectedCurrency := util.EUR

	// Add the expected SQL query and result to the mock
	mock.ExpectExec("INSERT INTO accounts").
		WithArgs(expectedID, expectedBalance, expectedCurrency).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new account to be posted
	acct := account.Account{
		ID:       "1111",
		Balance:  expectedBalance,
		Currency: expectedCurrency,
	}

	// Call the method being tested
	result, err := d.PostAccount(context.Background(), acct)

	// Assert the expected result
	assert.NoError(t, err)
	expectedResult := account.Account{
		ID:        expectedID,
		Balance:   expectedBalance,
		Currency:  expectedCurrency,
		CreatedAt: result.CreatedAt,
	}
	assert.Equal(t, expectedResult, result)

	// Assert that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create a new Database instance with the mock DB connection
	d := Database{Client: db}

	id := "1111"

	// Expect the query to be executed and return a successful result
	mock.ExpectExec("DELETE FROM accounts where id = ?").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the DeleteAccount method
	err = d.DeleteAccount(context.Background(), id)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTransaction(t *testing.T) {
	// Create a new SQL mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create SQL mock: %v", err)
	}
	defer db.Close()

	// Create a new Database instance with the mock DB connection
	d := Database{Client: db}

	// Define the expected transaction and row data
	expectedID := "1111"
	expectedSourceAccountID := "2222"
	expectedTargetAccountID := "3333"
	expectedAmount := 100.50
	expectedCurrency := util.EUR

	// Add the expected SQL query and result to the mock
	mock.ExpectQuery("SELECT id, source_account_id, target_account_id, amount, currency").
		WithArgs(expectedID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_account_id", "target_account_id",
			"amount", "currency"}).
			AddRow(expectedID, expectedSourceAccountID, expectedTargetAccountID, expectedAmount,
				expectedCurrency))

	// Call the method being tested
	result, err := d.GetTransaction(context.Background(), expectedID)

	// Assert the expected result
	assert.NoError(t, err)
	expectedResult := transaction.Transaction{
		ID:              expectedID,
		SourceAccountID: expectedSourceAccountID,
		TargetAccountID: expectedTargetAccountID,
		Amount:          expectedAmount,
		Currency:        expectedCurrency,
	}
	assert.Equal(t, expectedResult, result)

	// Assert that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransfer(t *testing.T) {
	// Create a new SQL mock
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	// Create a new Database instance using the mock DB
	d := &Database{Client: db}

	// Set up the test data
	ctx := context.Background()
	txn := transaction.Transaction{
		ID:              "1234",
		SourceAccountID: "1111",
		TargetAccountID: "2222",
		Amount:          100,
		Currency:        util.EUR,
	}
	sacc := account.Account{
		ID:        "1111",
		Balance:   500,
		Currency:  util.EUR,
		CreatedAt: time.Now(),
	}
	tacc := account.Account{
		ID:        "2222",
		Balance:   200,
		Currency:  util.EUR,
		CreatedAt: time.Now(),
	}

	// Set up the expected transaction Begin and Commit calls
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(sacc.Balance, sacc.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE accounts SET balance =").
		WithArgs(tacc.Balance, tacc.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Set up the expected insert query and result for the transaction
	mock.ExpectExec("INSERT INTO transactions").
		WithArgs(txn.ID, txn.SourceAccountID, txn.TargetAccountID, txn.Amount, txn.Currency).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	// Call the Transfer method
	result, err := d.Transfer(ctx, txn, sacc, tacc)
	assert.NoError(t, err)
	assert.Equal(t, txn, result)

	// Verify that all expectations were met
	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)

	// Set up the expected query and result for the source account
	mock.ExpectQuery("SELECT id, balance, currency, created_at FROM accounts WHERE id =").
		WithArgs(sacc.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "balance", "currency", "created_at"}).
			AddRow(sacc.ID, sacc.Balance-txn.Amount, sacc.Currency, sacc.CreatedAt))

	// Set up the expected query and result for the target account
	mock.ExpectQuery("SELECT id, balance, currency, created_at FROM accounts WHERE id =").
		WithArgs(tacc.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "balance", "currency", "created_at"}).
			AddRow(tacc.ID, tacc.Balance+txn.Amount, tacc.Currency, tacc.CreatedAt))

	// Verify the expected balance for the source account
	expectedSrcBalance := sacc.Balance - txn.Amount
	actualSrcBalance, err := d.GetAccount(ctx, sacc.ID)
	assert.NoError(t, err)
	assert.Equal(t, expectedSrcBalance, actualSrcBalance.Balance)

	// Verify the expected balance for the target account
	expectedTrgBalance := tacc.Balance + txn.Amount
	actualTrgBalance, err := d.GetAccount(ctx, tacc.ID)
	assert.NoError(t, err)
	assert.Equal(t, expectedTrgBalance, actualTrgBalance.Balance)
}

func TestDeleteTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create a new Database instance with the mock DB connection
	d := Database{Client: db}

	id := "1111"

	// Expect the query to be executed and return a successful result
	mock.ExpectExec("DELETE FROM transactions where id = ?").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Call the DeleteTransaction method
	err = d.DeleteTransaction(context.Background(), id)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Ensure all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}
