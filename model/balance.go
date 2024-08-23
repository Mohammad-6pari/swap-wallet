package model

type Balance struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	CryptoID  int     `json:"crypto_id"`
	Balance   int64 `json:"balance"`
}