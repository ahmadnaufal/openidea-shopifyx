CREATE TABLE IF NOT EXISTS bank_accounts (
  id VARCHAR(64) PRIMARY KEY,
  user_id VARCHAR(64) NOT NULL,
  bank_name VARCHAR(16) NOT NULL,
  bank_account_name VARCHAR(16) NOT NULL,
  bank_account_number VARCHAR(16) NOT NULL,
  created_at TIMESTAMP(0) DEFAULT NOW(),
  updated_at TIMESTAMP(0) DEFAULT NOW(),
  deleted_at TIMESTAMP(0)
);

CREATE INDEX IF NOT EXISTS idx_bank_accounts_user_id ON bank_accounts(user_id);
