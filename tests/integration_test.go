//go:build integration
// +build integration

package tests

import (
	"context"
	"financial-app/internals/database"
	"financial-app/pkg/account"
	"financial-app/util/validation"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccountDatabase(t *testing.T) {
	t.Run("test create account", func(t *testing.T) {
		db, err := database.NewDatabase()
		assert.NoError(t, err)

		acct, err := db.PostAccount(context.Background(), account.Account{
			ID:       "8de96b25-4c7a-4073-87da-e7b21c9308e1",
			Balance:  11.50,
			Currency: validation.EUR,
		})
		assert.NoError(t, err)

		newAcct, err := db.GetAccount(context.Background(), acct.ID)
		assert.NoError(t, err)
		assert.Equal(t, 11.50, newAcct.Balance)
		assert.Equal(t, validation.EUR, newAcct.Currency)

		// Clean account
		db.DeleteAccount(context.Background(), "8de96b25-4c7a-4073-87da-e7b21c9308e1")
	})

	t.Run("test delete account", func(t *testing.T) {
		db, err := database.NewDatabase()
		assert.NoError(t, err)
		acct, err := db.PostAccount(context.Background(), account.Account{
			ID:       "8de96b25-4c7a-4073-87da-e7b21c9308e1",
			Balance:  12.15,
			Currency: validation.EUR,
		})
		assert.NoError(t, err)

		err = db.DeleteAccount(context.Background(), acct.ID)
		assert.NoError(t, err)

		_, err = db.GetAccount(context.Background(), acct.ID)
		assert.Error(t, err)
	})

	t.Run("test get account", func(t *testing.T) {
		db, err := database.NewDatabase()
		assert.NoError(t, err)
		acct, err := db.PostAccount(context.Background(), account.Account{
			ID:       "8de96b25-4c7a-4073-87da-e7b21c9308e1",
			Balance:  12.50,
			Currency: validation.EUR,
		})
		assert.NoError(t, err)

		newAcct, err := db.GetAccount(context.Background(), acct.ID)
		assert.NoError(t, err)
		assert.Equal(t, 12.50, newAcct.Balance)
		assert.Equal(t, validation.EUR, newAcct.Currency)

		// Clean account
		db.DeleteAccount(context.Background(), "8de96b25-4c7a-4073-87da-e7b21c9308e1")
	})
}
