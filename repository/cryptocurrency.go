package repository

import (
	"database/sql"
	"swap-wallet/model"
)

type CryptocurrencyRepository struct {
	db *sql.DB
}

func NewCryptocurrencyRepository(db *sql.DB) *CryptocurrencyRepository {
	return &CryptocurrencyRepository{
		db: db,
	}
}

func (r *CryptocurrencyRepository) FindBySymbol(symbol string) (int ,error) {
	query := `SELECT id, name, symbol, is_available, scale FROM cryptocurrencies WHERE symbol = $1`

	var crypto model.Cryptocurrency
	err := r.db.QueryRow(query, symbol).Scan(
		&crypto.ID,
		&crypto.Name,
		&crypto.Symbol,
		&crypto.IsAvailable,
		&crypto.Scale,
	)
	if err == sql.ErrNoRows {
		return -1, nil
	} else if err != nil {
		return -1, err
	}

	return crypto.ID, nil
}

func (r *CryptocurrencyRepository) GetCryptoScale(symbol string) (int, error) {
	query := `SELECT scale FROM cryptocurrencies WHERE symbol = $1`

	var scale int
	err := r.db.QueryRow(query, symbol).Scan(&scale)
	if err != nil {
		return -1, err
	}

	return scale, nil
}