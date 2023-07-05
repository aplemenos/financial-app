//go:build e2e
// +build e2e

package tests

import (
	"encoding/json"
	"financial-app/pkg/models"
	"fmt"
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

	req, err := http.NewRequest("POST", BASE_URL+"/api/v1/account", strings.NewReader(acctBody))
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

	a := models.Account{}
	err = json.NewDecoder(rsp.Body).Decode(&a)
	if err != nil {
		return "", err
	}

	return a.ID, nil
}

func cleanAccount(id string) error {
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", BASE_URL+"/api/v1/account/"+id, nil)
	if err != nil {
		return err
	}

	req.Close = true
	req.Header.Add("Connection", "close")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	rsp.Body.Close()

	return nil
}

func cleanTransaction(id string) error {
	client := &http.Client{}

	req, err := http.NewRequest("DELETE", BASE_URL+"/api/v1/transaction/"+id, nil)
	if err != nil {
		return err
	}

	req.Close = true
	req.Header.Add("Connection", "close")

	rsp, err := client.Do(req)
	if err != nil {
		return err
	}

	rsp.Body.Close()

	return nil
}

func TestPostTransactionHappyPath(t *testing.T) {
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

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transaction",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	txn := models.Transaction{}
	err = json.NewDecoder(txnRsp.Body).Decode(&txn)
	assert.NoError(t, err)

	assert.Equal(t, 200, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)

	err = cleanAccount(taID)
	assert.NoError(t, err)

	err = cleanTransaction(txn.ID)
	assert.NoError(t, err)
}

func TestPostTransactionInsufficientBalance(t *testing.T) {
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

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transaction",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, 409, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)

	err = cleanAccount(taID)
	assert.NoError(t, err)
}

func TestPostTransactionSameAccount(t *testing.T) {
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

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transaction",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, 409, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)
}

func TestPostTransactionAccountNotFound(t *testing.T) {
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

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transaction",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, 404, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)
}

func TestPostTransactionEmptyAccount(t *testing.T) {
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

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transaction",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, 400, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)
}

func TestPostTransactionZeroAmount(t *testing.T) {
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

	txnReq, err := http.NewRequest("POST", BASE_URL+"/api/v1/transaction",
		strings.NewReader(txnBody))
	assert.NoError(t, err)

	txnReq.Close = true
	txnReq.Header.Add("Connection", "close")

	txnRsp, err := client.Do(txnReq)
	assert.NoError(t, err)

	defer txnRsp.Body.Close()

	assert.Equal(t, 400, txnRsp.StatusCode)

	// Cleanups
	err = cleanAccount(saID)
	assert.NoError(t, err)

	err = cleanAccount(taID)
	assert.NoError(t, err)
}

func TestHealthEndpoint(t *testing.T) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", BASE_URL+"/alive", nil)
	assert.NoError(t, err)

	req.Close = true
	req.Header.Add("Connection", "close")

	resp, err := client.Do(req)
	assert.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
}
