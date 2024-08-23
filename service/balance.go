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

)

type BalanceService struct {
	balanceRepo *repository.BalanceRepository
}

func NewBalanceService(balanceRepo *repository.BalanceRepository) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
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