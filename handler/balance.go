package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "swap-wallet/service"
    "github.com/gorilla/mux"
)

type BalanceHandler struct {
    balanceService *service.BalanceService
}

func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
    return &BalanceHandler{balanceService: balanceService}
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)

    userID, err := strconv.Atoi(vars["userId"])
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    cryptoSymbol := vars["cryptoSymbol"]

    cryptoBalance, usdBalance, err := h.balanceService.GetUserBalance(userID, cryptoSymbol)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	response := map[string]interface{}{
		"symbol": cryptoSymbol,
		"cryptoBalance": cryptoBalance,
		"usdBalance":    usdBalance,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *BalanceHandler) GetAllUserBalances(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	userID, err := strconv.Atoi(vars["userId"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	balances, err := h.balanceService.GetUserBalances(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balances)
}