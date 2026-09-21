package db

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func Connect() (*pgxpool.Pool, error) {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Auto-migrate: ensure schema constraints and columns exist
	migrationQuery := `
		ALTER TABLE customers 
		ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255) NOT NULL DEFAULT '';

		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM pg_constraint WHERE conname = 'chk_accounts_balance_non_negative'
			) THEN
				ALTER TABLE accounts ADD CONSTRAINT chk_accounts_balance_non_negative CHECK (balance >= 0);
			END IF;
		END $$;
	`
	if _, err := pool.Exec(ctx, migrationQuery); err != nil {
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	return pool, nil
}
