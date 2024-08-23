package model

type Cryptocurrency struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	IsAvailable bool   `json:"is_available"`
	Scale        int    `json:"scale"`
}