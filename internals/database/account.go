package database

import (
	"context"
	"database/sql"
	"financial-app/pkg/models"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// AccountRow - models how our account look in the database
type AccountRow struct {
	ID        string
	Balance   float64
	Currency  string
	CreatedAt sql.NullTime
}

func convertAccountRowToAccount(a AccountRow) models.Account {
	return models.Account{
		ID:        a.ID,
		Balance:   a.Balance,
		Currency:  a.Currency,
		CreatedAt: a.CreatedAt.Time,
	}
}

// GetAccount- retrieves an account from the database by ID
func (d *Database) GetAccount(
	ctx context.Context, uuid string,
) (models.Account, error) {
	// Fetch AccountRow from the database and then convert to models.Account
	var acctRow AccountRow
	row := d.Client.QueryRowContext(
		ctx,
		`SELECT id, balance, currency, created_at 
		FROM accounts 
		WHERE id = $1`,
		uuid,
	)
	err := row.Scan(
		&acctRow.ID,
		&acctRow.Balance,
		&acctRow.Currency,
		&acctRow.CreatedAt)
	if err != nil {
		return models.Account{},
			fmt.Errorf("an error occurred fetching an account by uuid: %w", err)
	}

	return convertAccountRowToAccount(acctRow), nil
}

// PostAccount- adds a new account to the database
func (d *Database) PostAccount(
	ctx context.Context, acct models.Account,
) (models.Account, error) {
	acctRow := AccountRow{
		ID:       acct.ID,
		Balance:  acct.Balance,
		Currency: acct.Currency,
	}

	// Define the insert query
	query := "INSERT INTO accounts (id, balance, currency) VALUES ($1, $2, $3)"

	result, err := d.Client.ExecContext(
		ctx, query, acctRow.ID, acctRow.Balance, acctRow.Currency,
	)
	if err != nil {
		return models.Account{}, fmt.Errorf("failed to insert account: %w", err)
	}

	// Get the number of affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Account{}, fmt.Errorf("failed to get affected rows: %w", err)
	}
	log.Info("The total number of affected rows ", rowsAffected)

	return acct, nil
}

// DeleteAccount - deletes an account from the database
func (d *Database) DeleteAccount(ctx context.Context, id string) error {
	_, err := d.Client.ExecContext(
		ctx,
		`DELETE FROM accounts where id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete accounts from the database: %w", err)
	}
	return nil
}
