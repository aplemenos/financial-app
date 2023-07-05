package database

import (
	"context"
	"financial-app/pkg/models"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// TransactionRow - models how our transaction look in the database
type TransactionRow struct {
	ID              string
	SourceAccountID string `db:"source_account_id"`
	TargetAccountID string `db:"target_account_id"`
	Amount          float64
	Currency        string
}

func convertTransactionRowToTransaction(t TransactionRow) models.Transaction {
	return models.Transaction{
		ID:              t.ID,
		SourceAccountID: t.SourceAccountID,
		TargetAccountID: t.TargetAccountID,
		Amount:          t.Amount,
		Currency:        t.Currency,
	}
}

// GetTransaction - retrieves a transaction from the database by ID
func (d *Database) GetTransaction(
	ctx context.Context, uuid string,
) (models.Transaction, error) {
	// Fetch TransactionRow from the database and then convert to models.Transaction
	var txnRow TransactionRow

	row := d.Client.QueryRowContext(
		ctx,
		`SELECT id, source_account_id, target_account_id, amount, currency 
		FROM transactions 
		WHERE id = $1`,
		uuid,
	)
	err := row.Scan(
		&txnRow.ID,
		&txnRow.SourceAccountID,
		&txnRow.TargetAccountID,
		&txnRow.Amount,
		&txnRow.Currency)
	if err != nil {
		return models.Transaction{},
			fmt.Errorf("an error occurred fetching a transaction by uuid: %w", err)
	}

	return convertTransactionRowToTransaction(txnRow), nil
}

// PostTransaction - adds a new transaction to the database
func (d *Database) PostTransaction(
	ctx context.Context, txn models.Transaction,
) (models.Transaction, error) {
	postRow := TransactionRow{
		ID:              txn.ID,
		SourceAccountID: txn.SourceAccountID,
		TargetAccountID: txn.TargetAccountID,
		Amount:          txn.Amount,
		Currency:        txn.Currency,
	}

	result, err := d.Client.ExecContext(
		ctx,
		`INSERT INTO transactions 
		(id, source_account_id, target_account_id, amount, currency) VALUES
		($1, $2, $3, $4, $5)`,
		postRow.ID, postRow.SourceAccountID, postRow.TargetAccountID, postRow.Amount,
		postRow.Currency,
	)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to insert transaction: %w", err)
	}

	// Get the number of affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to get affected rows: %w", err)
	}
	log.Info("The total number of affected rows ", rowsAffected)

	return txn, nil
}

// DeleteTransaction - deletes a transaction from the database
func (d *Database) DeleteTransaction(ctx context.Context, id string) error {
	_, err := d.Client.ExecContext(
		ctx,
		`DELETE FROM transactions where id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to delete transaction from the database: %w", err)
	}
	return nil
}
