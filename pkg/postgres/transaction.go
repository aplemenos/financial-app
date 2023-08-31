package postgres

// Transaction models how our transaction look in the database
type Transaction struct {
	ID              string
	SourceAccountID string `db:"source_account_id"`
	TargetAccountID string `db:"target_account_id"`
	Amount          float64
	Currency        string
}
