package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"time"
	"swap-wallet/repository"
	"swap-wallet/config"
	"github.com/dgrijalva/jwt-go"
	"os"
)

type BalanceService struct {
	balanceRepo *repository.BalanceRepository
	cryptoRepo *repository.CryptocurrencyRepository
}

func NewBalanceService(balanceRepo *repository.BalanceRepository, cryptoRepo *repository.CryptocurrencyRepository) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		cryptoRepo: cryptoRepo,
	}
}

func (s *BalanceService) formatURL(cryptoSymbol string, toSymbol string) string {
	return fmt.Sprintf(config.CryptoCompareAPI, cryptoSymbol, toSymbol)
}

func (s *BalanceService) getCryptoPrice(cryptoSymbol string, sourceSymbol string) (float64, error) {
	url := s.formatURL(cryptoSymbol, sourceSymbol)

	client := &http.Client{
		Timeout: config.Timeout * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch cryptocurrency price: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code from CryptoCompare API: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return 0, fmt.Errorf("failed to unmarshal JSON response: %v", err)
	}

	_, priceNotExist := data["Response"].(string)
	if priceNotExist {
		originToUsd, errOriginToUsd := s.getCryptoPrice(cryptoSymbol, "USD")
		if errOriginToUsd != nil {
			return 0, fmt.Errorf("failed to extract price for origin price from response")
		}

		sourceToUsd, errSourceToUsd := s.getCryptoPrice(sourceSymbol, "USD")
		if errSourceToUsd != nil {
			return 0, fmt.Errorf("failed to extract price for source price from response")
		}
		return originToUsd / sourceToUsd, nil
	}

	raw, ok := data["RAW"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("failed to extract RAW data from response")
	}

	price, ok := raw["PRICE"].(float64)
	if !ok {
		return 0, fmt.Errorf("failed to extract PRICE from RAW data")
	}

	return price, nil
}

func (s *BalanceService) getCryptoPriceInUSD(cryptoSymbol string) (float64, error) {
	return s.getCryptoPrice(cryptoSymbol, "USD")
}

func (s *BalanceService) GetUserBalance(userID int, cryptoSymbol string) (float64, float64, error) {
	balance, scale, err := s.balanceRepo.GetBalanceAndScale(userID, cryptoSymbol)
	if err != nil {
		return 0, 0, err
	}

	divisor := math.Pow(10, float64(scale))
	adjustedBalance := float64(balance) / divisor

	price, err := s.getCryptoPriceInUSD(cryptoSymbol)
	if err != nil {
		return 0, 0, err
	}

	usdBalance := adjustedBalance * price

	return adjustedBalance, usdBalance, nil
}

func (s *BalanceService) GetUserBalances(userID int) (map[string]map[string]float64, error) {
	balances, err := s.balanceRepo.GetAllBalancesForUser(userID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]float64)

	for _, balance := range balances {
		price, err := s.getCryptoPriceInUSD(balance.Symbol)
		if err != nil {
			return nil, fmt.Errorf("failed to get price for %s: %v", balance.Symbol, err)
		}

		divisor := math.Pow(10, float64(balance.Scale))
		adjustedBalance := float64(balance.Balance) / divisor
		usdBalance := adjustedBalance * price

		result[balance.Symbol] = map[string]float64{
			"cryptoBalance": adjustedBalance,
			"usdBalance":    usdBalance,
		}
	}

	return result, nil
}


func (s *BalanceService) GetConversionRate(sourceCrypto, targetCrypto string, amount float64) (float64, float64, error) {
	conversionRate, err := s.getCryptoPrice(sourceCrypto, targetCrypto)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get price for %s: %v", sourceCrypto, err)
	}
	convertedAmount := amount * conversionRate
	return conversionRate, convertedAmount, nil
}

func (s *BalanceService) FinalizeConversion(userID int, tokenString string) error {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return fmt.Errorf("invalid token: %v", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		conversionRate := claims["conversionRate"].(float64)
		sourceCrypto := claims["sourceCrypto"].(string)
		targetCrypto := claims["targetCrypto"].(string)
		amount := claims["amount"].(float64)

		balance, scale, err := s.balanceRepo.GetBalanceAndScale(userID, sourceCrypto)
		if err != nil {
			return fmt.Errorf("failed to get user balance: %v", err)
		}

		adjustedBalance := float64(balance) / math.Pow(10, float64(scale))
		if adjustedBalance < amount {
			return fmt.Errorf("insufficient balance")
		}

		sourceCryptoId, err := s.cryptoRepo.FindBySymbol(sourceCrypto)
		if err != nil {
			return fmt.Errorf("failed to find by symbol: %v", err)
		}

		newSourceBalance := adjustedBalance - amount
		err = s.balanceRepo.UpdateBalance(userID, sourceCryptoId, int64(newSourceBalance*math.Pow(10, float64(scale))))
		if err != nil {
			return fmt.Errorf("failed to update source balance: %v", err)
		}

		// Add to target crypto
		targetBalance, targetScale, err := s.balanceRepo.GetBalanceAndScale(userID, targetCrypto)
		if err != nil {
			return fmt.Errorf("failed to get target balance: %v", err)
		}

		targetCryptoId, err := s.cryptoRepo.FindBySymbol(targetCrypto)
		if err != nil {
			return fmt.Errorf("failed to find by symbol target: %v", err)
		}

		convertedAmount := amount * conversionRate
		newTargetBalance := (float64(targetBalance) / math.Pow(10, float64(targetScale))) + convertedAmount
		err = s.balanceRepo.UpdateBalance(userID, targetCryptoId, int64(newTargetBalance*math.Pow(10, float64(targetScale))))
		if err != nil {
			return fmt.Errorf("failed to update target balance: %v", err)
		}

		return nil
	} else {
		return fmt.Errorf("invalid token")
	}
}