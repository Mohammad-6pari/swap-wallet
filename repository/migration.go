package repository

import (
	"database/sql"
	"fmt"
	"swap-wallet/util"
)

func CreateTables(db *sql.DB) {
	userTable := `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) NOT NULL UNIQUE
		);`

	cryptoTable := `CREATE TABLE IF NOT EXISTS cryptocurrencies (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL UNIQUE,
		symbol VARCHAR(50) NOT NULL UNIQUE,
		is_available BOOLEAN NOT NULL,
		scale INT NOT NULL
	);`

	balanceTable := `CREATE TABLE IF NOT EXISTS balances (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		crypto_id INT NOT NULL,
		balance BIGINT NOT NULL,
		UNIQUE (user_id, crypto_id),
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (crypto_id) REFERENCES cryptocurrencies(id)
	);`

	_, err := db.Exec(userTable)
	util.CheckErr(err)
	fmt.Println("User table created or already exists.")

	_, err = db.Exec(cryptoTable)
	util.CheckErr(err)
	fmt.Println("Cryptocurrency table created or already exists.")

	_, err = db.Exec(balanceTable)
	util.CheckErr(err)
	fmt.Println("Balance table created or already exists.")
}
