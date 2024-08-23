package repository

import (
    "database/sql"
	"fmt"
)

type BalanceRepository struct {
    db *sql.DB
}

func NewBalanceRepository(db *sql.DB) *BalanceRepository {
    return &BalanceRepository{db: db}
}

func (r *BalanceRepository) GetBalanceAndScale(userID int, cryptoSymbol string) (int64, int, error) {
	var balance int64
	var scale int

	query := `
        SELECT COALESCE(b.balance, 0), c.scale
        FROM cryptocurrencies c
        LEFT JOIN balances b ON b.crypto_id = c.id AND b.user_id = $1
        WHERE c.symbol = $2
    `

	err := r.db.QueryRow(query, userID, cryptoSymbol).Scan(&balance, &scale)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, fmt.Errorf("cryptocurrency not found for symbol: %s", cryptoSymbol)
		}
		return -1, -1, err
	}

	return balance, scale, nil
}

func (r *BalanceRepository) GetAllBalancesForUser(userID int) ([]struct {
	CryptoID  int
	Symbol    string
	Balance   int64
	Scale     int
}, error) {
	query := `
		SELECT b.crypto_id, c.symbol, b.balance, c.scale
		FROM balances b
		JOIN cryptocurrencies c ON b.crypto_id = c.id
		WHERE b.user_id = $1
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch balances: %v", err)
	}
	defer rows.Close()

	var balances []struct {
		CryptoID  int
		Symbol    string
		Balance   int64
		Scale     int
	}
	for rows.Next() {
		var balance struct {
			CryptoID  int
			Symbol    string
			Balance   int64
			Scale     int
		}
		if err := rows.Scan(&balance.CryptoID, &balance.Symbol, &balance.Balance, &balance.Scale); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return balances, nil
}

func (r *BalanceRepository) GetBalanceForUser(userID int, cryptoID int) (int64, error) {
	var balance int64
	query := `SELECT balance FROM balances WHERE user_id = $1 AND crypto_id = $2`
	err := r.db.QueryRow(query, userID, cryptoID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch balance: %v", err)
	}
	return balance, nil
}

func (r *BalanceRepository) UpdateBalance(userID int, cryptoID int, newBalance int64) error {
	query := `
		INSERT INTO balances (user_id, crypto_id, balance)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, crypto_id)
		DO UPDATE SET balance = EXCLUDED.balance
	`

	_, err := r.db.Exec(query, userID, cryptoID, newBalance)
	if err != nil {
		return fmt.Errorf("failed to update or insert balance: %v", err)
	}

	return nil
}