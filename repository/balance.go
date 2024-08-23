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
        SELECT b.balance, c.scale
        FROM balances b
        JOIN cryptocurrencies c ON b.crypto_id = c.id
        WHERE b.user_id = $1 AND c.symbol = $2
    `

    err := r.db.QueryRow(query, userID, cryptoSymbol).Scan(&balance, &scale)
    if err != nil {
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