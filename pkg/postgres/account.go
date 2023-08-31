package postgres

import "database/sql"

// Account models how our account look in the database
type Account struct {
	ID        string
	Balance   float64
	Currency  string
	CreatedAt sql.NullTime
}
