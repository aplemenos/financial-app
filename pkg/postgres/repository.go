package postgres

import (
	"context"
	"database/sql"
	"financial-app/pkg/accounts"
	"financial-app/pkg/healthchecks"
	"financial-app/pkg/transactions"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/allisson/go-pglock/v3"
)

type accountRepository struct {
	client *sql.DB
	logger *zap.SugaredLogger
}

// NewAccountRepository returns a new instance of a postgres account repository.
func NewAccountRepository(
	client *sql.DB, logger *zap.SugaredLogger,
) accounts.AccountRepository {
	r := &accountRepository{
		client: client,
		logger: logger,
	}

	return r
}

func (r *accountRepository) Store(
	ctx context.Context, acct *accounts.Account,
) (*accounts.Account, error) {
	acctRow := Account{
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
		return nil, accounts.ErrPostingAccount(acct.ID)
	}

	return acct, nil
}

func convertAccountRowToAccount(a Account) *accounts.Account {
	return &accounts.Account{
		ID:        a.ID,
		Balance:   a.Balance,
		Currency:  a.Currency,
		CreatedAt: a.CreatedAt.Time,
	}
}

func (r *accountRepository) Find(
	ctx context.Context, id string,
) (*accounts.Account, error) {
	// Fetch accountRow from the database and then convert to Account
	var acctRow Account
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
		return nil, accounts.ErrFetchingAccount(id)

	}

	return convertAccountRowToAccount(acctRow), nil
}

func (r *accountRepository) FindByIDs(
	ctx context.Context, ids []string,
) (map[string]*accounts.Account, error) {
	// Do not process if the list is empty
	if len(ids) == 0 {
		return nil, accounts.ErrEmptyAccountList
	}

	// Create a map to store the fetched account rows by UUID
	acctRows := make(map[string]Account)

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
		return nil, accounts.ErrQueryingAccounts(ids)
	}
	defer rows.Close()

	// Iterate through the rows and store them in the map
	for rows.Next() {
		var acctRow Account
		err := rows.Scan(
			&acctRow.ID,
			&acctRow.Balance,
			&acctRow.Currency,
			&acctRow.CreatedAt,
		)
		if err != nil {
			r.logger.Errorf("an error occurred scanning account row: %w", err)
			return nil, accounts.ErrScanAccounts(ids)
		}
		acctRows[acctRow.ID] = acctRow
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		r.logger.Errorf("an error occurred iterating account rows: %w", err)
		return nil, accounts.ErrFetchingAccounts(ids)
	}

	// Convert the account rows to account
	accounts := make(map[string]*accounts.Account)
	for _, acctRow := range acctRows {
		accounts[acctRow.ID] = convertAccountRowToAccount(acctRow)
	}

	r.logger.Info(accounts)

	return accounts, nil
}

func (r *accountRepository) FindAll(
	ctx context.Context,
) []*accounts.Account {
	// Fetch all account rows from the database
	rows, err := r.client.QueryContext(
		ctx,
		`SELECT id, balance, currency, created_at 
		FROM accounts`,
	)
	if err != nil {
		r.logger.Errorf("an error occurred quering account rows:  %w", err)
		return []*accounts.Account{}
	}
	defer rows.Close()

	// Convert each accountRow to Account
	accts := make([]*accounts.Account, 0)
	for rows.Next() {
		var acctRow Account
		err := rows.Scan(
			&acctRow.ID,
			&acctRow.Balance,
			&acctRow.Currency,
			&acctRow.CreatedAt,
		)
		if err != nil {
			r.logger.Errorf("an error occurred scanning account row:  %w", err)
			return []*accounts.Account{}
		}
		acct := convertAccountRowToAccount(acctRow)
		accts = append(accts, acct)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("an error occurred iterating transaction rows: %w", err)
		return []*accounts.Account{}
	}

	return accts
}

func (r *accountRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.ExecContext(
		ctx,
		`DELETE FROM accounts where id = $1`,
		id,
	)
	if err != nil {
		r.logger.Errorf("failed to delete accounts from the database: %w", err)
		return accounts.ErrDeletingAccount(id)
	}
	return nil
}

type transactionRepository struct {
	client *sql.DB
	logger *zap.SugaredLogger
}

// NewTransactionRepository returns a new instance of a postgres transaction repository.
func NewTransactionRepository(
	client *sql.DB, logger *zap.SugaredLogger,
) transactions.TransactionRepository {
	r := &transactionRepository{
		client: client,
		logger: logger,
	}

	return r
}

func convertTransactionRowToTransaction(t Transaction) *transactions.Transaction {
	return &transactions.Transaction{
		ID:              t.ID,
		SourceAccountID: t.SourceAccountID,
		TargetAccountID: t.TargetAccountID,
		Amount:          t.Amount,
		Currency:        t.Currency,
	}
}

func (r *transactionRepository) Find(
	ctx context.Context, id string,
) (*transactions.Transaction, error) {
	// Fetch transactionRow from the database and then convert to Transaction
	var txnRow Transaction

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
		return nil, transactions.ErrFetchingTransaction(id)

	}

	return convertTransactionRowToTransaction(txnRow), nil
}

func (r *transactionRepository) FindAll(
	ctx context.Context,
) []*transactions.Transaction {
	// Fetch all transaction rows from the database
	rows, err := r.client.QueryContext(
		ctx,
		`SELECT id, source_account_id, target_account_id, amount, currency 
		FROM transactions`,
	)
	if err != nil {
		r.logger.Errorf("an error occurred quering transaction rows:  %w", err)
		return []*transactions.Transaction{}
	}
	defer rows.Close()

	// Convert each transactionRow to Transaction
	transacts := make([]*transactions.Transaction, 0)
	for rows.Next() {
		var txnRow Transaction
		err := rows.Scan(
			&txnRow.ID,
			&txnRow.SourceAccountID,
			&txnRow.TargetAccountID,
			&txnRow.Amount,
			&txnRow.Currency,
		)
		if err != nil {
			r.logger.Errorf("an error occurred scanning transaction row:  %w", err)
			return []*transactions.Transaction{}
		}
		txn := convertTransactionRowToTransaction(txnRow)
		transacts = append(transacts, txn)
	}

	if err = rows.Err(); err != nil {
		r.logger.Errorf("an error occurred iterating transaction rows: %w", err)
		return []*transactions.Transaction{}
	}

	return transacts
}

func (r *transactionRepository) Delete(
	ctx context.Context, id string,
) error {
	_, err := r.client.ExecContext(
		ctx,
		`DELETE FROM transactions where id = $1`,
		id,
	)
	if err != nil {
		r.logger.Errorf("failed to delete transaction from the database: %w", err)
		return transactions.ErrDeletingTransaction(id)
	}
	return nil
}

func (r *transactionRepository) Transfer(
	ctx context.Context,
	txn *transactions.Transaction,
	sacc *accounts.Account,
	tacc *accounts.Account,
) (*transactions.Transaction, error) {
	postRow := Transaction{
		ID:              txn.ID,
		SourceAccountID: txn.SourceAccountID,
		TargetAccountID: txn.TargetAccountID,
		Amount:          txn.Amount,
		Currency:        txn.Currency,
	}

	id := int64(1)
	lock, err := pglock.NewLock(ctx, id, r.client)
	if err != nil {
		r.logger.Errorf("failed to initialise the advisory lock in the db: %w", err)
		return nil, err
	}

	// Obtains exclusive session level advisory lock
	ok, err := lock.Lock(ctx)
	if err != nil {
		r.logger.Errorf("failed to lock the db: %w", err)
		return nil, err
	}
	r.logger.Info("lock.Lock()==", ok)

	// Release the lock
	defer func() {
		r.logger.Info("release lock")
		if err = lock.Unlock(ctx); err != nil {
			r.logger.Errorf("failed to unlock the db: %w", err)
		}
	}()

	// Transfer money securely from one account to another one through DB transactions
	err = r.executeDBTransaction(func(tx *sql.Tx) error {
		r.logger.Info("transfer ongoing...")
		// Define the update query for the accounts
		query := "UPDATE accounts SET balance = $1 WHERE id = $2"
		// Update the source account
		_, err := tx.ExecContext(ctx, query, sacc.Balance, sacc.ID)
		if err != nil {
			r.logger.Errorf("failed to update the source account: %w", err)
			return transactions.ErrUpdateAccount(sacc.ID)
		}

		// Update the target account
		_, err = tx.ExecContext(ctx, query, tacc.Balance, tacc.ID)
		if err != nil {
			r.logger.Errorf("failed to update the target account: %w", err)
			return transactions.ErrUpdateAccount(tacc.ID)
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
			return transactions.ErrPostingTransaction(txn.ID)
		}

		r.logger.Info("transfer completed")

		return nil
	})

	if err != nil {
		return nil, err
	}

	return convertTransactionRowToTransaction(postRow), nil
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

type healthcheckRepository struct {
	client *sql.DB
	logger *zap.SugaredLogger
}

// NewHealthcheckRepository returns a new instance of a postgres healthcheck repository.
func NewHealthcheckRepository(
	client *sql.DB, logger *zap.SugaredLogger,
) healthchecks.HealthcheckRepository {
	r := &healthcheckRepository{
		client: client,
		logger: logger,
	}

	return r
}

func (r *healthcheckRepository) Ping(ctx context.Context) error {
	return r.client.PingContext(ctx)
}
