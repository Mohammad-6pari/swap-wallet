package main

import (
	"net/http"
	"swap-wallet/config"
	handlers "swap-wallet/handler"
	"swap-wallet/repository"
	"swap-wallet/service"
	"swap-wallet/util"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.LoadConfig()

	db, err := util.ConnectDB(cfg)
	util.CheckErr(err)
	defer db.Close()

	redisClient, err := util.ConnectRedis(cfg)
	util.CheckErr(err)
	defer redisClient.Close()

	repository.CreateTables(db)
	// consider this method just run one time
	// repository.SeedPostgresData(db)

	balanceRepo := repository.NewBalanceRepository(db)
	cryptoRepo := repository.NewCryptocurrencyRepository(db)
	userRepo := repository.NewUserRepository(db)
	balanceService := service.NewBalanceService(balanceRepo, cryptoRepo, userRepo, redisClient)
	balanceHandler := handlers.NewBalanceHandler(balanceService)

	router := mux.NewRouter()
	router.HandleFunc("/balance", balanceHandler.GetUserBalance).Methods("GET")
	router.HandleFunc("/balances", balanceHandler.GetAllUserBalances).Methods("GET")
	router.HandleFunc("/exchange/preview", balanceHandler.GetExchangePreviewHandler).Methods("GET")
	router.HandleFunc("/exchange/apply", balanceHandler.FinalizeExchangeHandler).Methods("POST")
	http.ListenAndServe(":8080", router)

}
