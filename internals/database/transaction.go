package database

import (
	"context"
	"financial-app/pkg/models"
	"fmt"

	uuid "github.com/satori/go.uuid"
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
	txn.ID = uuid.NewV4().String()
	postRow := TransactionRow{
		ID:              txn.ID,
		SourceAccountID: txn.SourceAccountID,
		TargetAccountID: txn.TargetAccountID,
		Amount:          txn.Amount,
		Currency:        txn.Currency,
	}

	rows, err := d.Client.NamedQueryContext(
		ctx,
		`INSERT INTO transactions 
		(id, source_account_id, target_account_id, amount, currency) VALUES
		(:id, :source_account_id, :target_account_id, :amount, :currency)`,
		postRow,
	)
	if err != nil {
		return models.Transaction{}, fmt.Errorf("failed to insert transaction: %w", err)
	}
	if err := rows.Close(); err != nil {
		return models.Transaction{}, fmt.Errorf("failed to close rows: %w", err)
	}

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
