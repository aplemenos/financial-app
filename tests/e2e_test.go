//go:build e2e
// +build e2e

package tests

import (
	"encoding/json"
	"errors"
	"financial-app/pkg/account"
	"financial-app/pkg/transaction"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	BASE_URL = "http://localhost:8080"
)

func createAccount(amount float64) (string, error) {
	client := &http.Client{}

	acctBody := `{
		"balance": ` + fmt.Sprintf("%v", amount) + `,
		"currency": "EUR"}`

	req, err := http.NewRequest("POST", BASE_URL+"/api/v1/accounts", strings.NewReader(acctBody))
	if err != nil {
		return "", err
	}

	req.Close = true
	req.Header.Add("Connection", "close")

	rsp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		err, _ := io.ReadAll(rsp.Body)
		return "", errors.New("failed to create account: " + string(err))
	}

	a := account.Account{}
	err = json.NewDecoder(rsp.Body).Decode(&a)
	if err != nil {
		return "", err
	}

	return a.ID, nil
}

func cleanAccount(id string) error {
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", BASE_URL+"/api/v1/accounts/"+id, nil)
	if err != nil {
		return err
	}

	req.Close = true
	req.Header.Add("Connection", "close")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusNoContent {
		err, _ := io.ReadAll(rsp.Body)
		return errors.New("failed to delete account: " + string(err))
	}

	rsp.Body.Close()

	return nil
}

func cleanTransaction(id string) error {
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", BASE_URL+"/api/v1/transactions/"+id, nil)
	if err != nil {
		return err
	}

	req.Close = true
	req.Header.Add("Connection", "close")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusNoContent {
		err, _ := io.ReadAll(rsp.Body)
		return errors.New("failed to delete account: " + string(err))
	}

	rsp.Body.Close()

	return nil
}

func TestE2E_TransactionHappyPath(t *testing.T) {
	client := &http.Client{}

	// Create the source account
	saID, err := createAccount(11.50)
	assert.NoError(t, err)

	// Create the target account
	taID, err := createAccount(10.50)
	assert.NoError(t, err)

	// Perform a transaction
	txnBody := `{
		"source_account_id": "` + saID + `", 
		"target_account_id": "` + taID + `", 
		"amount": 11.50,
		"currency": "EUR"}`

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transactions",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	txn := transaction.Transaction{}
	err = json.NewDecoder(txnRsp.Body).Decode(&txn)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)

	err = cleanAccount(taID)
	assert.NoError(t, err)

	err = cleanTransaction(txn.ID)
	assert.NoError(t, err)
}

func TestE2E_TransactionInsufficientBalance(t *testing.T) {
	client := &http.Client{}

	// Create the source account
	saID, err := createAccount(9.50)
	assert.NoError(t, err)

	// Create the target account
	taID, err := createAccount(10.50)
	assert.NoError(t, err)

	// Perform a transaction
	txnBody := `{
		"source_account_id": "` + saID + `", 
		"target_account_id": "` + taID + `", 
		"amount": 11.50,
		"currency": "EUR"}`

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transactions",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, http.StatusConflict, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)

	err = cleanAccount(taID)
	assert.NoError(t, err)
}

func TestE2E_TransactionSameAccount(t *testing.T) {
	client := &http.Client{}

	// Create the source account
	saID, err := createAccount(11.50)
	assert.NoError(t, err)

	// Perform a transaction
	txnBody := `{
		"source_account_id": "` + saID + `", 
		"target_account_id": "` + saID + `", 
		"amount": 11.50,
		"currency": "EUR"}`

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transactions",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, http.StatusConflict, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)
}

func TestE2E_TransactionAccountNotFound(t *testing.T) {
	client := &http.Client{}

	// Create the source account
	saID, err := createAccount(11.50)
	assert.NoError(t, err)

	// Perform a transaction
	txnBody := `{
		"source_account_id": "` + saID + `", 
		"target_account_id": "11111111-1111-1111-1111-111111111111",
		"amount": 11.50,
		"currency": "EUR"}`

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transactions",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, http.StatusNotFound, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)
}

func TestE2E_TransactionEmptyAccount(t *testing.T) {
	client := &http.Client{}

	// Create the source account
	saID, err := createAccount(11.50)
	assert.NoError(t, err)

	// Perform a transaction
	txnBody := `{
		"source_account_id": "` + saID + `", 
		"target_account_id": "", 
		"amount": 11.50,
		"currency": "EUR"}`

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transactions",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)
}

func TestE2E_TransactionZeroAmount(t *testing.T) {
	client := &http.Client{}

	// Create the source account
	saID, err := createAccount(11.50)
	assert.NoError(t, err)

	// Create the target account
	taID, err := createAccount(10.50)
	assert.NoError(t, err)

	// Perform a transaction
	txnBody := `{
		"source_account_id": "` + saID + `", 
		"target_account_id": "` + taID + `", 
		"amount": 0,
		"currency": "EUR"}`

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transactions",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)

	err = cleanAccount(taID)
	assert.NoError(t, err)
}

func TestE2E_HealthEndpoint(t *testing.T) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", BASE_URL+"/alive", nil)
	assert.NoError(t, err)

	req.Close = true
	req.Header.Add("Connection", "close")

	resp, err := client.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
