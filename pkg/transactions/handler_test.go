package transactions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) Load(ctx context.Context, id string) (Transaction, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Transaction), args.Error(1)
}

func (m *MockService) LoadAll(ctx context.Context) []Transaction {
	args := m.Called(ctx)
	return args.Get(0).([]Transaction)
}

func (m *MockService) Transfer(ctx context.Context, txn Transaction) (Transaction, error) {
	args := m.Called(ctx, txn)
	return args.Get(0).(Transaction), args.Error(1)
}

func (m *MockService) Clean(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestTransactionHandler_Load(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	handler := &TransactionHandler{Service: mockService, Logger: logger.Sugar()}

	r := gin.Default()
	r.GET("/transactions/:id", handler.load)

	testCases := []struct {
		Name          string
		TransactionID string
		ExpectedError error
		ExpectedCode  int
	}{
		{
			Name:          "Transaction Found",
			TransactionID: "valid-transaction-id",
			ExpectedError: nil,
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "Transaction Not Found",
			TransactionID: "non-existent-transaction-id",
			ExpectedError: ErrFetchingTransaction("non-existent-transaction-id"),
			ExpectedCode:  http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockService.On("Load", mock.Anything, tc.TransactionID).
				Return(Transaction{}, tc.ExpectedError)

			req, _ := http.NewRequest("GET", "/transactions/"+tc.TransactionID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.ExpectedCode, rr.Code)

			// If an error is expected, assert the error message
			if tc.ExpectedError != nil {
				// Convert the JSON response to a map
				var response map[string]string
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				// Grab the value & whether or not it exists
				errorMsg, exists := response["error"]
				// Make some assertions on the correctness of the response.
				assert.Nil(t, err)
				assert.True(t, exists)
				assert.Equal(t, tc.ExpectedError.Error(), errorMsg)
			}
		})
	}
}

func TestTransactionHandler_LoadAll(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	handler := &TransactionHandler{Service: mockService, Logger: logger.Sugar()}

	r := gin.Default()
	r.GET("/transactions", handler.loadAll)

	testCases := []struct {
		Name             string
		ExpectedResponse []Transaction
		ExpectedCode     int
	}{
		{
			Name: "Transactions Found",
			ExpectedResponse: []Transaction{
				{ID: "transaction-id-1", SourceAccountID: "account-id-1",
					TargetAccountID: "account-id-2", Amount: 100, Currency: "EUR"},
				{ID: "transaction-id-2", SourceAccountID: "account-id-3",
					TargetAccountID: "account-id-4", Amount: 200, Currency: "EUR"},
			},
			ExpectedCode: http.StatusOK,
		},
		{
			Name:             "No Transactions Found",
			ExpectedResponse: nil,
			ExpectedCode:     http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockService.On("LoadAll", mock.Anything).
				Return(tc.ExpectedResponse)

			req, _ := http.NewRequest("GET", "/transactions", nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.ExpectedCode, rr.Code)

			// If transactions are expected, assert the transactions in the response body
			if tc.ExpectedResponse != nil {
				var transactions []Transaction
				err := json.NewDecoder(rr.Body).Decode(&transactions)
				assert.NoError(t, err)
				assert.Equal(t, tc.ExpectedResponse, transactions)
			}
		})
	}
}

func TestTransactionHandler_Clean(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	handler := &TransactionHandler{Service: mockService, Logger: logger.Sugar()}

	r := gin.Default()
	r.DELETE("/transactions/:id", handler.clean)

	testCases := []struct {
		Name          string
		TransactionID string
		ExpectedError error
		ExpectedCode  int
	}{
		{
			Name:          "Transaction Deleted",
			TransactionID: "valid-transaction-id",
			ExpectedError: nil,
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "Transaction Not Deleted",
			TransactionID: "transaction-id-not-deleted",
			ExpectedError: ErrDeletingTransaction("non-existent-transaction-id"),
			ExpectedCode:  http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockService.On("Clean", mock.Anything, tc.TransactionID).
				Return(tc.ExpectedError)

			req, _ := http.NewRequest("DELETE", "/transactions/"+tc.TransactionID, nil)
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.ExpectedCode, rr.Code)

			// If an error is expected, assert the error message
			if tc.ExpectedError != nil {
				// Convert the JSON response to a map
				var response map[string]string
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				// Grab the value & whether or not it exists
				errorMsg, exists := response["error"]
				// Make some assertions on the correctness of the response.
				assert.Nil(t, err)
				assert.True(t, exists)
				assert.Equal(t, tc.ExpectedError.Error(), errorMsg)
			}
		})
	}
}

func TestTransactionHandler_Transfer(t *testing.T) {
	mockService := new(MockService)
	logger, _ := zap.NewDevelopment()
	handler := &TransactionHandler{Service: mockService, Logger: logger.Sugar()}

	r := gin.Default()
	r.POST("/transactions", handler.transfer)

	testCases := []struct {
		Name             string
		Request          transactionRequest
		ExpectedError    error
		ExpectedCode     int
		ExpectedResponse Transaction
	}{
		{
			Name: "Valid Transfer",
			Request: transactionRequest{
				SourceAccountID: "4067bfcb-d722-4e0e-a15e-b16be3b00f84",
				TargetAccountID: "fd27c968-7a4a-4bcf-b880-8aa93e5ab2d1",
				Amount:          100,
				Currency:        "USD",
			},
			ExpectedError: nil,
			ExpectedCode:  http.StatusOK,
			ExpectedResponse: Transaction{
				ID:              "967c2536-57ed-410a-bb2e-08a002e73138",
				SourceAccountID: "4067bfcb-d722-4e0e-a15e-b16be3b00f84",
				TargetAccountID: "fd27c968-7a4a-4bcf-b880-8aa93e5ab2d1",
				Amount:          100,
				Currency:        "USD",
			},
		},
		{
			Name: "Invalid Currency",
			Request: transactionRequest{
				SourceAccountID: "4067bfcb-d722-4e0e-a15e-b16be3b00f84",
				TargetAccountID: "fd27c968-7a4a-4bcf-b880-8aa93e5ab2d1",
				Amount:          200,
				Currency:        "XYZ",
			},
			ExpectedError:    errors.New(currencyNotSupported),
			ExpectedCode:     http.StatusBadRequest,
			ExpectedResponse: Transaction{},
		},
		{
			Name: "Accounts Same",
			Request: transactionRequest{
				SourceAccountID: "4067bfcb-d722-4e0e-a15e-b16be3b00f84",
				TargetAccountID: "4067bfcb-d722-4e0e-a15e-b16be3b00f84",
				Amount:          400,
				Currency:        "USD",
			},
			ExpectedError:    errors.New(accountsNotSame),
			ExpectedCode:     http.StatusConflict,
			ExpectedResponse: Transaction{},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mockService.On("Transfer", mock.Anything, mock.Anything).
				Return(tc.ExpectedResponse, tc.ExpectedError)

			requestBody, _ := json.Marshal(tc.Request)
			req, _ := http.NewRequest("POST", "/transactions", bytes.NewReader(requestBody))
			rr := httptest.NewRecorder()

			r.ServeHTTP(rr, req)

			assert.Equal(t, tc.ExpectedCode, rr.Code)

			// If a response is expected, assert the response in the body
			if tc.ExpectedError != nil {
				// Convert the JSON response to a map
				var response map[string]string
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				// Grab the value & whether or not it exists
				errorMsg, exists := response["error"]
				// Make some assertions on the correctness of the response.
				assert.Nil(t, err)
				assert.True(t, exists)
				assert.Equal(t, tc.ExpectedError.Error(), errorMsg)
			}
		})
	}
}
