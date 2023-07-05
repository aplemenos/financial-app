package database

import (
	"context"
	"database/sql"
	"financial-app/pkg/models"
	"testing"

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
	expectedCurrency := "EUR"
	expectedCreatedAt := sql.NullTime{Valid: true}

	// Add the expected SQL query and result to the mock
	mock.ExpectQuery("SELECT id, balance, currency, created_at").
		WithArgs(expectedID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "balance", "currency", "created_at"}).
			AddRow(expectedID, expectedBalance, expectedCurrency, expectedCreatedAt))

	// Call the method being tested
	account, err := d.GetAccount(context.Background(), expectedID)

	// Assert the expected result
	assert.NoError(t, err)
	expectedAccount := models.Account{
		ID:        expectedID,
		Balance:   expectedBalance,
		Currency:  expectedCurrency,
		CreatedAt: expectedCreatedAt.Time,
	}
	assert.Equal(t, expectedAccount, account)

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
	expectedCurrency := "EUR"

	// Add the expected SQL query and result to the mock
	mock.ExpectExec("INSERT INTO accounts").
		WithArgs(expectedID, expectedBalance, expectedCurrency).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new account to be posted
	account := models.Account{
		ID:       "1111",
		Balance:  expectedBalance,
		Currency: expectedCurrency,
	}

	// Call the method being tested
	result, err := d.PostAccount(context.Background(), account)

	// Assert the expected result
	assert.NoError(t, err)
	expectedResult := models.Account{
		ID:        expectedID,
		Balance:   expectedBalance,
		Currency:  expectedCurrency,
		CreatedAt: result.CreatedAt,
	}
	assert.Equal(t, expectedResult, result)

	// Assert that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateAccount(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Create a new Database instance with the mock DB connection
	d := Database{Client: db}

	// Define the expected account and row data
	expectedID := "1111"
	expectedBalance := 100.0
	expectedCurrency := "EUR"

	// Expect the query to be executed and return the mock rows
	mock.ExpectExec("UPDATE accounts SET").
		WithArgs(expectedBalance, expectedCurrency, expectedID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Call the UpdateAccount method
	account := models.Account{
		ID:       expectedID,
		Balance:  expectedBalance,
		Currency: expectedCurrency,
	}
	updatedAccount, err := d.UpdateAccount(context.Background(), expectedID, account)

	// Assert that no error occurred
	assert.NoError(t, err)

	// Assert the updated account values
	expectedAccount := models.Account{
		ID:       expectedID,
		Balance:  expectedBalance,
		Currency: expectedCurrency,
	}
	assert.Equal(t, expectedAccount, updatedAccount)

	// Ensure all expectations were met
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
