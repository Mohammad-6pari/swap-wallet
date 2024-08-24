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

type FinalizeRequest struct {
	UserID int    `json:"userId"`
	Token  string `json:"token"`
}

func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
    return &BalanceHandler{balanceService: balanceService}
}

func (h *BalanceHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query()
	crypto := query.Get("crypto")

    userID, err := strconv.Atoi(r.Header.Get("userId"))
    
	if err != nil {
        http.Error(w, "User ID Should Be Int", http.StatusBadRequest)
        return
    }

	userExists := h.balanceService.UserExists(userID)
	if !userExists {
		http.Error(w, "Invalid Uesr ID", http.StatusNotFound)
        return
	}

    cryptoBalance, usdBalance, err := h.balanceService.GetUserBalanceWithUsd(userID, crypto)
	if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

	response := map[string]interface{}{
		"crypto": crypto,
		"cryptoBalance": cryptoBalance,
		"USDBalance":    usdBalance,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *BalanceHandler) GetAllUserBalances(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.Header.Get("userId"))
	
	if err != nil {
        http.Error(w, "User ID Should Be Int", http.StatusBadRequest)
		return
	}

	userExists := h.balanceService.UserExists(userID)
	if !userExists {
		http.Error(w, "Invalid Uesr ID", http.StatusNotFound)
        return
	}

	balances, err := h.balanceService.GetUserBalancesWithUsd(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balances)
}

func (h *BalanceHandler) GetExchangePreviewHandler(w http.ResponseWriter, r *http.Request) {
	sourceAmountStr := r.URL.Query().Get("sourceAmount")
	source := r.URL.Query().Get("source")
	target := r.URL.Query().Get("target")
	if sourceAmountStr == "" || source == "" || target == "" {
		http.Error(w, "Missing query parameters", http.StatusBadRequest)
		return
	}

	sourceAmount, err := strconv.ParseFloat(sourceAmountStr, 64)
	if err != nil {
		http.Error(w, "Invalid sourceAmount", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.Header.Get("userId"))
	
	if err != nil {
        http.Error(w, "User ID Should Be Int", http.StatusBadRequest)
		return
	}

	userExists := h.balanceService.UserExists(userID)
	if !userExists {
		http.Error(w, "Invalid Uesr ID", http.StatusNotFound)
        return
	}

	convertedAmount, token, err := h.balanceService.GetExchangePreview(source, target, sourceAmount)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"convertedAmount": convertedAmount,
		"token":           token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *BalanceHandler) FinalizeConversionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userIDStr := vars["userId"]

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var requestData struct {
		Token string `json:"token"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.balanceService.FinalizeConversion(userID, requestData.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Conversion finalized successfully",
	})
}