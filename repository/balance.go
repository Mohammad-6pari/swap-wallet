package repository

import (
    "database/sql"
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
