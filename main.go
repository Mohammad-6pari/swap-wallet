package main

import (
	"swap-wallet/config"
	"swap-wallet/repository"
	"swap-wallet/util"
	"swap-wallet/handler"
    "swap-wallet/service"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	cfg := config.LoadConfig()

	db, err := util.ConnectDB(cfg)
	util.CheckErr(err)
	defer db.Close()

	repository.CreateTables(db)
	// consider this method just run one time
	// repository.SeedPostgresData(db)
	
	balanceRepo := repository.NewBalanceRepository(db)
	cryptoRepo := repository.NewCryptocurrencyRepository(db)
	userRepo := repository.NewUserRepository(db)
    balanceService := service.NewBalanceService(balanceRepo, cryptoRepo, userRepo)
    balanceHandler := handlers.NewBalanceHandler(balanceService)

    router := mux.NewRouter()
    router.HandleFunc("/balance", balanceHandler.GetUserBalance).Methods("GET")
    router.HandleFunc("/balances", balanceHandler.GetAllUserBalances).Methods("GET")
	router.HandleFunc("/balance/{amount}/{sourceCrypto}/{targetCrypto}", balanceHandler.GetConversionRateHandler).Methods("GET")
	router.HandleFunc("/balance/convert/{userId}", balanceHandler.FinalizeConversionHandler).Methods("POST")
    http.ListenAndServe(":8080", router)

}
