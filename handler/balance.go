package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "swap-wallet/service"
    "github.com/gorilla/mux"
	"time"
	"os"
	"github.com/joho/godotenv"
	"github.com/dgrijalva/jwt-go"

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

func (h *BalanceHandler) GetConversionRateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	amountStr := vars["amount"]
	sourceCrypto := vars["sourceCrypto"]
	targetCrypto := vars["targetCrypto"]

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	conversionRate, convertedAmount, err := h.balanceService.GetConversionRate(sourceCrypto, targetCrypto, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

		
	err = godotenv.Load()
	if err != nil {
		http.Error(w, "Error loading .env file", http.StatusInternalServerError)
		return
	}
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	// Create a JWT token with the necessary information
	claims := jwt.MapClaims{
		"exp":         time.Now().Add(60 * time.Second).Unix(),
		"conversionRate": conversionRate,
		"sourceCrypto":   sourceCrypto,
		"targetCrypto":   targetCrypto,
		"amount":         amount,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"conversionRate": conversionRate,
		"convertedAmount": convertedAmount,
		"token": tokenString,
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