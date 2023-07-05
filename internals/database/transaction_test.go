package database

import (
	"context"
	"financial-app/pkg/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

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
	expectedCurrency := "EUR"

	// Add the expected SQL query and result to the mock
	mock.ExpectQuery("SELECT id, source_account_id, target_account_id, amount, currency").
		WithArgs(expectedID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "source_account_id", "target_account_id",
			"amount", "currency"}).
			AddRow(expectedID, expectedSourceAccountID, expectedTargetAccountID, expectedAmount,
				expectedCurrency))

	// Call the method being tested
	transaction, err := d.GetTransaction(context.Background(), expectedID)

	// Assert the expected result
	assert.NoError(t, err)
	expectedResult := models.Transaction{
		ID:              expectedID,
		SourceAccountID: expectedSourceAccountID,
		TargetAccountID: expectedTargetAccountID,
		Amount:          expectedAmount,
		Currency:        expectedCurrency,
	}
	assert.Equal(t, expectedResult, transaction)

	// Assert that all expectations were met
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostTransaction(t *testing.T) {
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
	expectedCurrency := "EUR"

	// Add the expected SQL query and result to the mock
	mock.ExpectExec("INSERT INTO transactions").
		WithArgs(expectedID, expectedSourceAccountID, expectedTargetAccountID,
			expectedAmount, expectedCurrency).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a new account to be posted
	transaction := models.Transaction{
		ID:              expectedID,
		SourceAccountID: expectedSourceAccountID,
		TargetAccountID: expectedTargetAccountID,
		Amount:          expectedAmount,
		Currency:        expectedCurrency,
	}

	// Call the method being tested
	result, err := d.PostTransaction(context.Background(), transaction)

	// Assert the expected result
	assert.NoError(t, err)
	expectedResult := models.Transaction{
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
