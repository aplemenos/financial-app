CREATE TABLE IF NOT EXISTS transactions (
    id uuid PRIMARY KEY,
    source_account_id uuid,
    target_account_id uuid,
    amount NUMERIC(8, 2),
    currency TEXT
);

CREATE TABLE IF NOT EXISTS accounts (
    id uuid PRIMARY KEY,
    balance NUMERIC(8, 2),
    currency TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);