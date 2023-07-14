package postgres

import (
	"context"
	"database/sql"
	"financial-app/domain/account"
	"financial-app/domain/transaction"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type accountRepository struct {
	client *sql.DB
	logger *zap.SugaredLogger
}

// accountRow models how our account look in the database
type accountRow struct {
	ID        string
	Balance   float64
	Currency  string
	CreatedAt sql.NullTime
}

// NewAccountRepository returns a new instance of a postgres account repository.
func NewAccountRepository(
	client *sql.DB, logger *zap.SugaredLogger,
) account.AccountRepository {
	r := &accountRepository{
		client: client,
		logger: logger,
	}

	return r
}

func (r *accountRepository) Store(
	ctx context.Context, acct *account.Account,
) (*account.Account, error) {
	acctRow := accountRow{
		ID:       string(acct.ID),
		Balance:  acct.Balance,
		Currency: string(acct.Currency),
	}

	// Define the insert query
	query := "INSERT INTO accounts (id, balance, currency) VALUES ($1, $2, $3)"

	_, err := r.client.ExecContext(
		ctx, query, acctRow.ID, acctRow.Balance, acctRow.Currency,
	)
	if err != nil {
		// fmt.Errorf("failed to insert account: %w", err)
		return nil, account.ErrPostingAccount
	}

	return acct, nil
}

func convertAccountRowToAccount(a accountRow) *account.Account {
	return &account.Account{
		ID:        account.AccountID(a.ID),
		Balance:   a.Balance,
		Currency:  account.Currency(a.Currency),
		CreatedAt: a.CreatedAt.Time,
	}
}

func (r *accountRepository) Find(
	ctx context.Context, id account.AccountID,
) (*account.Account, error) {
	// Fetch accountRow from the database and then convert to Account
	var acctRow accountRow
	row := r.client.QueryRowContext(
		ctx,
		`SELECT id, balance, currency, created_at 
		FROM accounts 
		WHERE id = $1`,
		id,
	)
	err := row.Scan(
		&acctRow.ID,
		&acctRow.Balance,
		&acctRow.Currency,
		&acctRow.CreatedAt)
	if err != nil {
		return nil, account.ErrFetchingAccount(id)

	}

	return convertAccountRowToAccount(acctRow), nil
}

func (r *accountRepository) FindByIDs(
	ctx context.Context, ids []account.AccountID,
) (map[account.AccountID]*account.Account, error) {
	// Do not process if the list is empty
	if len(ids) == 0 {
		return nil, account.ErrEmptyAccountList
	}

	// Create a map to store the fetched account rows by UUID
	acctRows := make(map[string]accountRow)

	// Build the query placeholders for the IN operator
	placeholders := make([]interface{}, len(ids))
	inquery := "$1"
	for i, id := range ids {
		placeholders[i] = id
		if i > 0 {
			inquery += ",$" + fmt.Sprint(i+1)
		}
	}

	// Execute the query and retrieve the account rows
	rows, err := r.client.QueryContext(
		ctx,
		`SELECT id, balance, currency, created_at
		FROM accounts
		WHERE id IN (`+inquery+`)`,
		placeholders...,
	)
	if err != nil {
		r.logger.Errorf("an error occurred fetching accounts by UUIDs: %w", err)
		return nil, account.ErrQueryingAccounts
	}
	defer rows.Close()

	// Iterate through the rows and store them in the map
	for rows.Next() {
		var acctRow accountRow
		err := rows.Scan(
			&acctRow.ID,
			&acctRow.Balance,
			&acctRow.Currency,
			&acctRow.CreatedAt,
		)
		if err != nil {
			r.logger.Errorf("an error occurred scanning account row: %w", err)
			return nil, account.ErrScanAccount
		}
		acctRows[acctRow.ID] = acctRow
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		r.logger.Errorf("an error occurred iterating account rows: %w", err)
		return nil, account.ErrFetchingAccounts
	}

	// Convert the account rows to account
	accounts := make(map[account.AccountID]*account.Account)
	for _, acctRow := range acctRows {
		accounts[account.AccountID(acctRow.ID)] = convertAccountRowToAccount(acctRow)
	}

	r.logger.Info(accounts)

	return accounts, nil
}

func (r *accountRepository) FindAll(
	ctx context.Context,
) []*account.Account {
	// Fetch all account rows from the database
	rows, err := r.client.QueryContext(
		ctx,
		`SELECT id, balance, currency, created_at 
		FROM accounts`,
	)
	if err != nil {
		r.logger.Errorf("an error occurred quering account rows:  %w", err)
		return []*account.Account{}
	}
	defer rows.Close()

	// Convert each accountRow to Account
	accounts := make([]*account.Account, 0)
	for rows.Next() {
		var acctRow accountRow
		err := rows.Scan(
			&acctRow.ID,
			&acctRow.Balance,
			&acctRow.Currency,
			&acctRow.CreatedAt,
		)
		if err != nil {
			r.logger.Errorf("an error occurred scanning account row:  %w", err)
			return []*account.Account{}
		}
		acct := convertAccountRowToAccount(acctRow)
		accounts = append(accounts, acct)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("an error occurred iterating transaction rows: %w", err)
		return []*account.Account{}
	}

	return accounts
}

func (r *accountRepository) Delete(ctx context.Context, id account.AccountID) error {
	_, err := r.client.ExecContext(
		ctx,
		`DELETE FROM accounts where id = $1`,
		id,
	)
	if err != nil {
		r.logger.Errorf("failed to delete accounts from the database: %w", err)
		return account.ErrDeletingAccount(id)
	}
	return nil
}

type transactionRepository struct {
	client *sql.DB
	logger *zap.SugaredLogger
}

// transactionRow models how our transaction look in the database
type transactionRow struct {
	ID              string
	SourceAccountID string `db:"source_account_id"`
	TargetAccountID string `db:"target_account_id"`
	Amount          float64
	Currency        string
}

// NewTransactionRepository returns a new instance of a postgres transaction repository.
func NewTransactionRepository(
	client *sql.DB, logger *zap.SugaredLogger,
) transaction.TransactionRepository {
	r := &transactionRepository{
		client: client,
		logger: logger,
	}

	return r
}

func convertTransactionRowToTransaction(t transactionRow) *transaction.Transaction {
	return &transaction.Transaction{
		ID:              transaction.TransactionID(t.ID),
		SourceAccountID: account.AccountID(t.SourceAccountID),
		TargetAccountID: account.AccountID(t.TargetAccountID),
		Amount:          t.Amount,
		Currency:        account.Currency(t.Currency),
	}
}

func (r *transactionRepository) Find(
	ctx context.Context, id transaction.TransactionID,
) (*transaction.Transaction, error) {
	// Fetch transactionRow from the database and then convert to Transaction
	var txnRow transactionRow

	row := r.client.QueryRowContext(
		ctx,
		`SELECT id, source_account_id, target_account_id, amount, currency 
		FROM transactions 
		WHERE id = $1`,
		id,
	)
	err := row.Scan(
		&txnRow.ID,
		&txnRow.SourceAccountID,
		&txnRow.TargetAccountID,
		&txnRow.Amount,
		&txnRow.Currency)
	if err != nil {
		return nil, transaction.ErrFetchingTransaction(id)

	}

	return convertTransactionRowToTransaction(txnRow), nil
}

func (r *transactionRepository) FindAll(
	ctx context.Context,
) []*transaction.Transaction {
	// Fetch all transaction rows from the database
	rows, err := r.client.QueryContext(
		ctx,
		`SELECT id, source_account_id, target_account_id, amount, currency 
		FROM transactions`,
	)
	if err != nil {
		r.logger.Errorf("an error occurred quering transaction rows:  %w", err)
		return []*transaction.Transaction{}
	}
	defer rows.Close()

	// Convert each transactionRow to Transaction
	transactions := make([]*transaction.Transaction, 0)
	for rows.Next() {
		var txnRow transactionRow
		err := rows.Scan(
			&txnRow.ID,
			&txnRow.SourceAccountID,
			&txnRow.TargetAccountID,
			&txnRow.Amount,
			&txnRow.Currency,
		)
		if err != nil {
			r.logger.Errorf("an error occurred scanning transaction row:  %w", err)
			return []*transaction.Transaction{}
		}
		txn := convertTransactionRowToTransaction(txnRow)
		transactions = append(transactions, txn)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("an error occurred iterating transaction rows: %w", err)
		return []*transaction.Transaction{}
	}

	return transactions
}

func (r *transactionRepository) Delete(
	ctx context.Context, id transaction.TransactionID,
) error {
	_, err := r.client.ExecContext(
		ctx,
		`DELETE FROM transactions where id = $1`,
		id,
	)
	if err != nil {
		r.logger.Errorf("failed to delete transaction from the database: %w", err)
		return transaction.ErrDeletingTransaction(id)
	}
	return nil
}

func (r *transactionRepository) Transfer(
	ctx context.Context, txn *transaction.Transaction, sacc *account.Account, tacc *account.Account,
) (*transaction.Transaction, error) {
	postRow := transactionRow{
		ID:              string(txn.ID),
		SourceAccountID: string(txn.SourceAccountID),
		TargetAccountID: string(txn.TargetAccountID),
		Amount:          txn.Amount,
		Currency:        string(txn.Currency),
	}

	// Transfer money securely from one account to another one through DB transactions
	err := r.executeDBTransaction(func(tx *sql.Tx) error {
		// Define the update query for the accounts
		query := "UPDATE accounts SET balance = $1 WHERE id = $2"
		// Update the source account
		_, err := tx.ExecContext(ctx, query, sacc.Balance, sacc.ID)
		if err != nil {
			r.logger.Errorf("failed to update the source account: %w", err)
			return transaction.ErrUpdateAccount(sacc.ID)
		}

		// Update the target account
		_, err = tx.ExecContext(ctx, query, tacc.Balance, tacc.ID)
		if err != nil {
			r.logger.Errorf("failed to update the target account: %w", err)
			return transaction.ErrUpdateAccount(tacc.ID)
		}

		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO transactions 
		(id, source_account_id, target_account_id, amount, currency) VALUES
		($1, $2, $3, $4, $5)`,
			postRow.ID, postRow.SourceAccountID, postRow.TargetAccountID, postRow.Amount,
			postRow.Currency,
		)
		if err != nil {
			r.logger.Errorf("failed to insert transaction: %w", err)
			return transaction.ErrPostingTransaction
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return convertTransactionRowToTransaction(postRow), nil
}

func (r *transactionRepository) Ping(ctx context.Context) error {
	return r.client.PingContext(ctx)
}

// executeDBTransaction executes a safe transaction via the provided function
func (r *transactionRepository) executeDBTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := r.client.Begin()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}
