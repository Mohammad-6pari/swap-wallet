package util

import (
	"database/sql"
	"fmt"
	"log"

	"swap-wallet/config"

	_ "github.com/lib/pq"
)

func ConnectDB(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetDBConnStr())
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to the database!")
	return db, nil
}

func CheckErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
