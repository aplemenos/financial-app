package accounts

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

// MockService is a mock implementation of the Service interface for testing.
type MockService struct {
	mock.Mock
}

func (m *MockService) Accounts(ctx context.Context) []Account {
	args := m.Called(ctx)
	return args.Get(0).([]Account)
}

func (m *MockService) LoadAccount(ctx context.Context, id string) (Account, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Account), args.Error(1)
}

func (m *MockService) Register(ctx context.Context, acct Account) (Account, error) {
	args := m.Called(ctx, acct)
	return args.Get(0).(Account), args.Error(1)
}

func (m *MockService) Clean(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestAccountHandler_LoadAll(t *testing.T) {
	// Create a mock service and an AccountHandler instance using the mock service
	logger, _ := zap.NewDevelopment()
	mockService := new(MockService)
	handler := &AccountHandler{Service: mockService, Logger: logger.Sugar()}

	// Create a router and register the accounts route
	r := gin.Default()
	r.GET("/accounts", handler.loadAll)

	// Define test cases
	testCases := []struct {
		Name             string
		ExpectedResponse []Account
		ExpectedCode     int
	}{
		{
			Name: "Accounts Found",
			ExpectedResponse: []Account{
				{ID: "account-id-1", Balance: 100, Currency: "EUR"},
				{ID: "account-id-2", Balance: 200, Currency: "EUR"},
			},
			ExpectedCode: http.StatusOK,
		},
		{
			Name:             "No Accounts",
			ExpectedResponse: nil,
			ExpectedCode:     http.StatusOK,
		},
	}

	// Iterate through test cases and run the tests
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Set up expected behavior for the mock service method
			mockService.On("Accounts", mock.Anything).Return(tc.ExpectedResponse)

			// Create a new HTTP request
			req, _ := http.NewRequest("GET", "/accounts", nil)
			rr := httptest.NewRecorder()

			// Serve the request using the router
			r.ServeHTTP(rr, req)

			// Perform assertions
			assert.Equal(t, tc.ExpectedCode, rr.Code) // Check HTTP response status code

			// If accounts are expected, assert the accounts in the response body
			if tc.ExpectedResponse != nil {
				var accounts []Account
				err := json.NewDecoder(rr.Body).Decode(&accounts)
				assert.NoError(t, err)
				assert.Equal(t, tc.ExpectedResponse, accounts)
			}
		})
	}
}

func TestAccountHandler_Register(t *testing.T) {
	// Create a mock service and an AccountHandler instance using the mock service
	logger, _ := zap.NewDevelopment()
	mockService := new(MockService)
	handler := &AccountHandler{Service: mockService, Logger: logger.Sugar()}

	// Create a router and register the registerAccount route
	r := gin.Default()
	r.POST("/accounts", handler.register)

	// Define test cases
	testCases := []struct {
		Name             string
		Request          storeRequest
		ExpectedError    error
		ExpectedCode     int
		ExpectedResponse Account
	}{
		{
			Name: "Valid Registration",
			Request: storeRequest{
				Balance:  100,
				Currency: "USD",
			},
			ExpectedError: nil,
			ExpectedCode:  http.StatusCreated,
			ExpectedResponse: Account{
				Balance:  100,
				Currency: "USD",
			},
		},
		{
			Name: "Invalid Currency",
			Request: storeRequest{
				Balance:  200,
				Currency: "XYZ",
			},
			ExpectedError:    errors.New(currencyNotSupported),
			ExpectedCode:     http.StatusBadRequest,
			ExpectedResponse: Account{},
		},
	}

	// Iterate through test cases and run the tests
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Set up expected behavior for the mock service method
			mockService.On("Register", mock.Anything, mock.Anything).Return(tc.ExpectedResponse,
				tc.ExpectedError)

			// Create a new HTTP request with the test case's request data
			requestBody, _ := json.Marshal(tc.Request)
			req, _ := http.NewRequest("POST", "/accounts", bytes.NewReader(requestBody))
			rr := httptest.NewRecorder()

			// Serve the request using the router
			r.ServeHTTP(rr, req)

			// Perform assertions
			assert.Equal(t, tc.ExpectedCode, rr.Code) // Check HTTP response status code

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

func TestAccountHandler_Load(t *testing.T) {
	// Create a mock service and an AccountHandler instance using the mock service
	logger, _ := zap.NewDevelopment()
	mockService := new(MockService)
	handler := &AccountHandler{Service: mockService, Logger: logger.Sugar()}

	// Create a router and register the loadAccount route
	r := gin.Default()
	r.GET("/accounts/:id", handler.load)

	// Define test cases
	testCases := []struct {
		Name          string
		AccountID     string
		ExpectedError error
		ExpectedCode  int
	}{
		{
			Name:          "Account Found",
			AccountID:     "valid-account-id",
			ExpectedError: nil,
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "Account Not Found",
			AccountID:     "non-existent-account-id",
			ExpectedError: ErrFetchingAccount("non-existent-account-id"),
			ExpectedCode:  http.StatusNotFound,
		},
	}

	// Iterate through test cases and run the tests
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Set up expected behavior for the mock service method
			mockService.On("LoadAccount", mock.Anything, tc.AccountID).
				Return(Account{}, tc.ExpectedError)

			// Create a new HTTP request
			req, _ := http.NewRequest("GET", "/accounts/"+tc.AccountID, nil)
			rr := httptest.NewRecorder()

			// Serve the request using the router
			r.ServeHTTP(rr, req)

			// Perform assertions
			assert.Equal(t, tc.ExpectedCode, rr.Code) // Check HTTP response status code

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

func TestAccountHandler_Clean(t *testing.T) {
	// Create a mock service and an AccountHandler instance using the mock service
	logger, _ := zap.NewDevelopment()
	mockService := new(MockService)
	handler := &AccountHandler{Service: mockService, Logger: logger.Sugar()}

	// Create a router and register the clean route
	r := gin.Default()
	r.DELETE("/accounts/:id", handler.clean)

	// Define test cases
	testCases := []struct {
		Name          string
		AccountID     string
		ExpectedError error
		ExpectedCode  int
	}{
		{
			Name:          "Account Deleted",
			AccountID:     "valid-account-id",
			ExpectedError: nil,
			ExpectedCode:  http.StatusOK,
		},
		{
			Name:          "Account Not Deleted",
			AccountID:     "account-id-not-deleted",
			ExpectedError: ErrDeletingAccount("account-id-not-deleted"),
			ExpectedCode:  http.StatusInternalServerError,
		},
	}

	// Iterate through test cases and run the tests
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Set up expected behavior for the mock service method
			mockService.On("Clean", mock.Anything, tc.AccountID).Return(tc.ExpectedError)

			// Create a new HTTP request
			req, _ := http.NewRequest("DELETE", "/accounts/"+tc.AccountID, nil)
			rr := httptest.NewRecorder()

			// Serve the request using the router
			r.ServeHTTP(rr, req)

			// Perform assertions
			assert.Equal(t, tc.ExpectedCode, rr.Code) // Check HTTP response status code

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
