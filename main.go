package main

import (
	"swap-wallet/config"
	"swap-wallet/repository"
	"swap-wallet/util"
)

func main() {
	cfg := config.LoadConfig()

	db, err := util.ConnectDB(cfg)
	util.CheckErr(err)
	defer db.Close()

	repository.CreateTables(db)
	repository.SeedPostgresData(db)
	
}
