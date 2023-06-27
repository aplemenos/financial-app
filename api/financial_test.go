package api

import (
	"bytes"
	"encoding/json"
	"financial-app/internals/transaction/repo/kvs"
	"financial-app/pkg/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTransactionHappyPath(t *testing.T) {
	// Initialize the server
	router := FinancialAppV1()

	// Create mock source and target accounts
	// Create two mock account data
	sourceAccount := models.Account{
		ID:        "54462360-67e2-47e6-9962-21ba7ec7f141",
		Balance:   1000.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}

	targetAccount := models.Account{
		ID:        "61b72db1-c001-4bf1-9a72-b2d6c6c8d8bd",
		Balance:   500.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}

	// Save them in the store
	a := kvs.NewAccountStore()
	a.Set(sourceAccount.ID, sourceAccount)
	a.Set(targetAccount.ID, targetAccount)

	// Cleanup accounts when test completed
	defer a.Delete(sourceAccount.ID)
	defer a.Delete(targetAccount.ID)

	// Prepare a transaction request payload
	transactionRequest := models.Transaction{
		SourceAccountID: sourceAccount.ID,
		TargetAccountID: targetAccount.ID,
		Amount:          200.0,
	}

	// Convert transaction request payload to JSON
	requestBody, _ := json.Marshal(transactionRequest)

	// Create a new HTTP POST request
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP response recorder
	recorder := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusCreated, recorder.Code)

	// Check the updated balances of the source and target accounts
	source, _ := a.Get(sourceAccount.ID)
	target, _ := a.Get(targetAccount.ID)
	assert.Equal(t, 800.0, source.Balance)
	assert.Equal(t, 700.0, target.Balance)
}

func TestTransactionInsufficientBalance(t *testing.T) {
	// Initialize the server
	router := FinancialAppV1()

	// Create mock source and target accounts
	// Create two mock account data
	sourceAccount := models.Account{
		ID:        "54462360-67e2-47e6-9962-21ba7ec7f141",
		Balance:   100.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}

	targetAccount := models.Account{
		ID:        "61b72db1-c001-4bf1-9a72-b2d6c6c8d8bd",
		Balance:   500.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}

	// Save them in the store
	a := kvs.NewAccountStore()
	a.Set(sourceAccount.ID, sourceAccount)
	a.Set(targetAccount.ID, targetAccount)

	// Cleanup accounts when test completed
	defer a.Delete(sourceAccount.ID)
	defer a.Delete(targetAccount.ID)

	// Prepare a transaction request payload
	transactionRequest := models.Transaction{
		SourceAccountID: sourceAccount.ID,
		TargetAccountID: targetAccount.ID,
		Amount:          200.0,
	}

	// Convert transaction request payload to JSON
	requestBody, _ := json.Marshal(transactionRequest)

	// Create a new HTTP POST request
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP response recorder
	recorder := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Check that the balances of the source and target accounts remain the same
	source, _ := a.Get(sourceAccount.ID)
	target, _ := a.Get(targetAccount.ID)
	assert.Equal(t, 100.0, source.Balance)
	assert.Equal(t, 500.0, target.Balance)
}

func TestTransactionSameAccount(t *testing.T) {
	// Initialize the server
	router := FinancialAppV1()

	// Create a mock account
	account := models.Account{
		ID:        "54462360-67e2-47e6-9962-21ba7ec7f141",
		Balance:   1000.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}
	// Save them in the store
	a := kvs.NewAccountStore()
	a.Set(account.ID, account)

	// Cleanup account when test completed
	defer a.Delete(account.ID)

	// Prepare a transaction request payload with the same source and target account ID
	transactionRequest := models.Transaction{
		SourceAccountID: account.ID,
		TargetAccountID: account.ID,
		Amount:          200.0,
	}

	// Convert transaction request payload to JSON
	requestBody, _ := json.Marshal(transactionRequest)

	// Create a new HTTP POST request
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP response recorder
	recorder := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Check that the balance of the account remains the same
	source, _ := a.Get(account.ID)
	assert.Equal(t, 1000.0, source.Balance)
}

func TestTransactionNonexistentAccount(t *testing.T) {
	// Initialize the server
	router := FinancialAppV1()

	// Create a mock account
	account := models.Account{
		ID:        "54462360-67e2-47e6-9962-21ba7ec7f141",
		Balance:   1000.0,
		Currency:  "EUR",
		CreatedAt: time.Now(),
	}
	// Save them in the store
	a := kvs.NewAccountStore()
	a.Set(account.ID, account)

	// Cleanup account when test completed
	defer a.Delete(account.ID)

	// Prepare a transaction request payload with a nonexistent target account ID
	transactionRequest := models.Transaction{
		SourceAccountID: account.ID,
		TargetAccountID: "61b72db1-c001-4bf1-9a72-b2d6c6c8d8bd",
		Amount:          200.0,
	}

	// Convert transaction request payload to JSON
	requestBody, _ := json.Marshal(transactionRequest)

	// Create a new HTTP POST request
	req := httptest.NewRequest("POST", "/transactions", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP response recorder
	recorder := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(recorder, req)

	// Check the response status code
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Check that the balance of the source account remains the same
	source, _ := a.Get(account.ID)
	assert.Equal(t, 1000.0, source.Balance)
}
