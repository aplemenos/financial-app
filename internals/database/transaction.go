package database

import (
	"context"
	"database/sql"
	"financial-app/pkg/account"
	"financial-app/pkg/transaction"
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

func convertAccountRowToAccount(a AccountRow) account.Account {
	return account.Account{
		ID:        a.ID,
		Balance:   a.Balance,
		Currency:  a.Currency,
		CreatedAt: a.CreatedAt.Time,
	}
}

// GetAccount- retrieves an account from the database by ID
func (d *Database) GetAccount(
	ctx context.Context, uuid string,
) (account.Account, error) {
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
		return account.Account{},
			fmt.Errorf("an error occurred fetching an account by uuid: %w", err)
	}

	return convertAccountRowToAccount(acctRow), nil
}

// PostAccount- adds a new account to the database
func (d *Database) PostAccount(
	ctx context.Context, acct account.Account,
) (account.Account, error) {
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
		return account.Account{}, fmt.Errorf("failed to insert account: %w", err)
	}

	// Get the number of affected rows
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return account.Account{}, fmt.Errorf("failed to get affected rows: %w", err)
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

// TransactionRow - models how our transaction look in the database
type TransactionRow struct {
	ID              string
	SourceAccountID string `db:"source_account_id"`
	TargetAccountID string `db:"target_account_id"`
	Amount          float64
	Currency        string
}

func convertTransactionRowToTransaction(t TransactionRow) transaction.Transaction {
	return transaction.Transaction{
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
) (transaction.Transaction, error) {
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
		return transaction.Transaction{},
			fmt.Errorf("an error occurred fetching a transaction by uuid: %w", err)
	}

	return convertTransactionRowToTransaction(txnRow), nil
}

// Transfer - performs a secure transaction from source to target account
func (d *Database) Transfer(
	ctx context.Context,
	txn transaction.Transaction,
	sacc account.Account,
	tacc account.Account,
) (transaction.Transaction, error) {
	postRow := TransactionRow{
		ID:              txn.ID,
		SourceAccountID: txn.SourceAccountID,
		TargetAccountID: txn.TargetAccountID,
		Amount:          txn.Amount,
		Currency:        txn.Currency,
	}

	err := d.ExecuteDBTransaction(func(tx *sql.Tx) error {
		// Define the update query for the accounts
		query := "UPDATE accounts SET balance = $1 WHERE id = $2"
		// Update the source account
		result, err := tx.ExecContext(ctx, query, sacc.Balance, sacc.ID)
		if err != nil {
			return fmt.Errorf("failed to update the source account: %w", err)
		}

		// Get the number of affected rows (source account)
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}
		log.Info("The affected rows of source account ", rowsAffected)

		// Update the target account
		result, err = tx.ExecContext(ctx, query, tacc.Balance, tacc.ID)
		if err != nil {
			return fmt.Errorf("failed to update the target account: %w", err)
		}

		// Get the number of affected rows (target account)
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}
		log.Info("The affected rows of target account ", rowsAffected)

		result, err = tx.ExecContext(
			ctx,
			`INSERT INTO transactions 
		(id, source_account_id, target_account_id, amount, currency) VALUES
		($1, $2, $3, $4, $5)`,
			postRow.ID, postRow.SourceAccountID, postRow.TargetAccountID, postRow.Amount,
			postRow.Currency,
		)
		if err != nil {
			return fmt.Errorf("failed to insert transaction: %w", err)
		}

		// Get the number of affected rows (transaction)
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get affected rows: %w", err)
		}
		log.Info("The affected rows of inserted transaction ", rowsAffected)
		return nil
	})

	if err != nil {
		return transaction.Transaction{}, err
	}

	return convertTransactionRowToTransaction(postRow), nil
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
