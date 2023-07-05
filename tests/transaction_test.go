//go:build integration
// +build integration

package tests

import (
	"context"
	"financial-app/internals/database"
	"financial-app/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionDatabase(t *testing.T) {
	t.Run("test create transaction", func(t *testing.T) {
		db, err := database.NewDatabase()
		assert.NoError(t, err)

		txn, err := db.PostTransaction(context.Background(), models.Transaction{
			ID:              "e0251039-a04d-447e-81d8-3e86e1e324b9",
			SourceAccountID: "7f083c6d-44b8-4e29-a0c8-0f7ae8892a91",
			TargetAccountID: "877a8a63-4791-41af-838e-2336d9e1a6a7",
			Amount:          11.50,
			Currency:        "EUR",
		})
		assert.NoError(t, err)

		newTxn, err := db.GetTransaction(context.Background(), txn.ID)
		assert.NoError(t, err)
		assert.Equal(t, "7f083c6d-44b8-4e29-a0c8-0f7ae8892a91", newTxn.SourceAccountID)
		assert.Equal(t, "877a8a63-4791-41af-838e-2336d9e1a6a7", newTxn.TargetAccountID)
		assert.Equal(t, 11.50, newTxn.Amount)
		assert.Equal(t, "EUR", newTxn.Currency)

		// Clean transaction
		db.DeleteTransaction(context.Background(), "e0251039-a04d-447e-81d8-3e86e1e324b9")
	})

	t.Run("test delete transaction", func(t *testing.T) {
		db, err := database.NewDatabase()
		assert.NoError(t, err)
		txn, err := db.PostTransaction(context.Background(), models.Transaction{
			ID:              "e0251039-a04d-447e-81d8-3e86e1e324b9",
			SourceAccountID: "7f083c6d-44b8-4e29-a0c8-0f7ae8892a91",
			TargetAccountID: "877a8a63-4791-41af-838e-2336d9e1a6a7",
			Amount:          11.50,
			Currency:        "EUR",
		})
		assert.NoError(t, err)

		err = db.DeleteTransaction(context.Background(), txn.ID)
		assert.NoError(t, err)

		_, err = db.GetTransaction(context.Background(), txn.ID)
		assert.Error(t, err)
	})

	t.Run("test get transaction", func(t *testing.T) {
		db, err := database.NewDatabase()
		assert.NoError(t, err)
		txn, err := db.PostTransaction(context.Background(), models.Transaction{
			ID:              "e0251039-a04d-447e-81d8-3e86e1e324b9",
			SourceAccountID: "7f083c6d-44b8-4e29-a0c8-0f7ae8892a91",
			TargetAccountID: "877a8a63-4791-41af-838e-2336d9e1a6a7",
			Amount:          11.50,
			Currency:        "EUR",
		})
		assert.NoError(t, err)

		newTxn, err := db.GetTransaction(context.Background(), txn.ID)
		assert.NoError(t, err)
		assert.Equal(t, "7f083c6d-44b8-4e29-a0c8-0f7ae8892a91", newTxn.SourceAccountID)
		assert.Equal(t, "877a8a63-4791-41af-838e-2336d9e1a6a7", newTxn.TargetAccountID)
		assert.Equal(t, 11.50, newTxn.Amount)
		assert.Equal(t, "EUR", newTxn.Currency)

		// Clean transaction
		db.DeleteTransaction(context.Background(), "e0251039-a04d-447e-81d8-3e86e1e324b9")
	})
}
