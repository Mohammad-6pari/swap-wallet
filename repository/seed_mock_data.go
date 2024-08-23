package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"swap-wallet/model"
	"swap-wallet/util"
)

func SeedPostgresData(db *sql.DB) {
	seedCryptocurrencies(db)
	seedUsers(db)
	seedBalances(db)

}

func seedCryptocurrencies(db *sql.DB) {
	file, err := os.Open("/app/data/cryptocurrency.json")
	util.CheckErr(err)
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	util.CheckErr(err)

	var cryptos []model.Cryptocurrency
	err = json.Unmarshal(byteValue, &cryptos)
	util.CheckErr(err)

	for _, crypto := range cryptos {
		_, err = db.Exec("INSERT INTO cryptocurrencies (name, symbol, is_available, scale) VALUES ($1, $2, $3, $4)",
			crypto.Name, crypto.Symbol, crypto.IsAvailable, crypto.Scale)
		util.CheckErr(err)
		fmt.Printf("Inserted cryptocurrency: %s\n", crypto.Name)
	}
}

func seedUsers(db *sql.DB) {
	file, err := os.Open("/app/data/users.json")
	util.CheckErr(err)
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	util.CheckErr(err)

	var users []model.User
	err = json.Unmarshal(byteValue, &users)
	util.CheckErr(err)

	for _, user := range users {
		_, err = db.Exec("INSERT INTO users (username) VALUES ($1)", user.Username)
		util.CheckErr(err)
		fmt.Printf("Inserted user: %s\n", user.Username)
	}
}

func seedBalances(db *sql.DB) {
	file, err := os.Open("/app/data/balances.json")
	util.CheckErr(err)
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	util.CheckErr(err)

	var balances []model.Balance
	err = json.Unmarshal(byteValue, &balances)
	util.CheckErr(err)

	for _, balance := range balances {
		_, err = db.Exec(
			"INSERT INTO balances (user_id, crypto_id, balance) VALUES ($1, $2, $3)",
			balance.UserID, balance.CryptoID, balance.Balance,
		)
		util.CheckErr(err)
		fmt.Printf("Inserted balance for user_id: %d, crypto_id: %d\n", balance.UserID, balance.CryptoID)
	}
}
