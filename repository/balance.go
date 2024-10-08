package repository

import (
	"database/sql"
	"fmt"
	"math"
)

type BalanceRepository struct {
	db *sql.DB
}

type CryptoBalance struct {
	CryptoName string `json:"crypto_name"`
	Balance    int64  `json:"balance"`
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

func (r *BalanceRepository) GetUserBalance(userID int, crypto string) (int64, error) {
	var balance int64
	query := `
		SELECT COALESCE(b.balance, 0)
		FROM balances b
		JOIN cryptocurrencies c ON b.crypto_id = c.id
		WHERE b.user_id = $1 AND c.symbol = $2
    `

	err := r.db.QueryRow(query, userID, crypto).Scan(&balance)
	if err != nil {
		return -1, err
	}

	return balance, nil
}

func (r *BalanceRepository) GetUserBalances(userID int) ([]CryptoBalance, error) {
	var userBalances []CryptoBalance

	query := `
        SELECT c.symbol, COALESCE(b.balance, 0)
        FROM balances b
        JOIN cryptocurrencies c ON b.crypto_id = c.id
        WHERE b.user_id = $1
    `

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ub CryptoBalance
		err := rows.Scan(&ub.CryptoName, &ub.Balance)
		if err != nil {
			return nil, err
		}
		userBalances = append(userBalances, ub)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userBalances, nil
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

func (r *BalanceRepository) UpdateBalance(tx *sql.Tx, userID int, cryptoSymbol string, newBalance int64) error {
	query := `
		INSERT INTO balances (user_id, crypto_id, balance)
		VALUES ($1, (SELECT id FROM cryptocurrencies WHERE symbol = $2), $3)
		ON CONFLICT (user_id, crypto_id)
		DO UPDATE SET balance = EXCLUDED.balance
	`
	_, err := tx.Exec(query, userID, cryptoSymbol, newBalance)
	if err != nil {
		return fmt.Errorf("failed to update or insert balance: %v", err)
	}

	return nil
}

func (r *BalanceRepository) ExchangeBalances(userID int, sourceCrypto, targetCrypto string, sourceAmount, targetAmount float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	sourceBalance, err := r.GetUserBalance(userID, sourceCrypto)
	if err != nil {
		return fmt.Errorf("failed to get source balance: %v", err)
	}

	sourceScale, err := r.GetCryptoScale(tx, sourceCrypto)
	if err != nil {
		return fmt.Errorf("failed to get source crypto scale: %v", err)
	}

	scaledSourceAmount := int64(sourceAmount * math.Pow(10, float64(sourceScale)))

	if sourceBalance < scaledSourceAmount {
		return fmt.Errorf("insufficient balance in source cryptocurrency")
	}

	newSourceBalance := sourceBalance - scaledSourceAmount
	err = r.UpdateBalance(tx, userID, sourceCrypto, newSourceBalance)
	if err != nil {
		return fmt.Errorf("failed to update source balance: %v", err)
	}

	targetBalance, err := r.GetUserBalance(userID, targetCrypto)
	if err != nil {
		return fmt.Errorf("failed to get target balance: %v", err)
	}

	targetScale, err := r.GetCryptoScale(tx, targetCrypto)
	if err != nil {
		return fmt.Errorf("failed to get target crypto scale: %v", err)
	}

	scaledTargetAmount := int64(targetAmount * math.Pow(10, float64(targetScale)))

	newTargetBalance := targetBalance + scaledTargetAmount
	err = r.UpdateBalance(tx, userID, targetCrypto, newTargetBalance)
	if err != nil {
		return fmt.Errorf("failed to update target balance: %v", err)
	}

	return nil
}
