package service

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"time"
	"swap-wallet/repository"
	"swap-wallet/config"
	"github.com/dgrijalva/jwt-go"
	"os"
	"log"
	"fmt"
	"github.com/joho/godotenv"
)

type BalanceService struct {
	balanceRepo *repository.BalanceRepository
	cryptoRepo *repository.CryptocurrencyRepository
	userRepo *repository.UserRepository
}

type CryptoBalanceType struct {
	CryptoName   string  `json:"crypto_name"`
	CryptoBalance float64 `json:"crypto_balance"`
	USDBalance    float64 `json:"usd_balance"`
}

func NewBalanceService(balanceRepo *repository.BalanceRepository, cryptoRepo *repository.CryptocurrencyRepository, userRepo *repository.UserRepository) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
		cryptoRepo: cryptoRepo,
		userRepo: userRepo,
	}
}

func (s *BalanceService) UserExists(userID int) (bool) {
	_, err := s.userRepo.GetUsername(userID)
	return err == nil
}

func formatURL(cryptoSymbol string, toSymbol string) string {
	return fmt.Sprintf(config.CryptoCompareAPI, cryptoSymbol, toSymbol)
}

func getCryptoPriceFromThirdParty(cryptoSymbol string, sourceSymbol string) (float64, error) {
	url := formatURL(cryptoSymbol, sourceSymbol)

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
		originToUsd, errOriginToUsd := getCryptoPriceFromThirdParty(cryptoSymbol, "USD")
		if errOriginToUsd != nil {
			return 0, fmt.Errorf("failed to extract price for origin price from response")
		}

		sourceToUsd, errSourceToUsd := getCryptoPriceFromThirdParty(sourceSymbol, "USD")
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

func (s *BalanceService) getUserBalance(userID int, crypto string) (float64, error) {
	balance, err := s.balanceRepo.GetUserBalance(userID, crypto)

	if err != nil {
		return 0, err
	}

	scale, err := s.cryptoRepo.GetCryptoScale(crypto)

	if err != nil {
		return 0, nil
	}
	
	adjustedCryptoBalance := scaleCryptoBalance(balance, scale)
	return adjustedCryptoBalance, nil
}

func scaleCryptoBalance(balance int64, scale int) (float64) {
	divisor := math.Pow(10, float64(scale))
	return float64(balance) / divisor
}

func (s *BalanceService) GetUserBalancesWithUsd(userID int) ([]CryptoBalanceType, error) {
	cryptoBalances, err := s.getUserBalances(userID)
	if err != nil {
		return nil, err
	}

	var result []CryptoBalanceType

	for _, cryptoBalance := range cryptoBalances {
		adjustedCryptoBalance, err := s.getUserBalance(userID, cryptoBalance.CryptoName)
		if err != nil {
			return nil, fmt.Errorf("failed to get balance for %s: %v", cryptoBalance.CryptoName, err)
		}

		price, err := getCryptoPriceFromThirdParty(cryptoBalance.CryptoName, "USD")
		if err != nil {
			return nil, fmt.Errorf("failed to get price for %s: %v", cryptoBalance.CryptoName, err)
		}

		usdBalance := adjustedCryptoBalance * price
		result = append(result, CryptoBalanceType{
			CryptoName:    cryptoBalance.CryptoName,
			CryptoBalance: adjustedCryptoBalance,
			USDBalance:    usdBalance,
		})
	}

	return result, nil
}

func (s *BalanceService) getUserBalances(userID int) ([]CryptoBalanceType, error) {
	cryptoBalances, err := s.balanceRepo.GetUserBalances(userID)
	if err != nil {
		return nil, err
	}

	var adjustedBalances []CryptoBalanceType

	for _, cryptoBalance := range cryptoBalances {
		scale, err := s.cryptoRepo.GetCryptoScale(cryptoBalance.CryptoName)
		if err != nil {
			log.Printf("Failed to get scale for %s: %v", cryptoBalance.CryptoName, err)
			continue
		}

		adjustedBalance := scaleCryptoBalance(cryptoBalance.Balance, scale)
		adjustedBalances = append(adjustedBalances, CryptoBalanceType{
			CryptoName: cryptoBalance.CryptoName,
			CryptoBalance:    adjustedBalance,
		})
	}

	return adjustedBalances, nil
}

func (s *BalanceService) GetUserBalanceWithUsd(userID int, crypto string) (float64, float64, error) {
	cryptoBalance, err := s.getUserBalance(userID, crypto)
	if err != nil {
		return 0, 0, err
	}

	crypoPriceUSDUnit, err := getCryptoPriceFromThirdParty(crypto, "USD")

	if err != nil {
		return 0, 0, err
	}

	UsdBalance := crypoPriceUSDUnit * cryptoBalance
	return cryptoBalance, UsdBalance, nil
}

func createJWTToken(source, target string, sourceAmount, targetAmount float64) (string, error) {
	err := godotenv.Load()
	if err != nil {
		return "", fmt.Errorf("error loading .env file: %v", err)
	}

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	claims := jwt.MapClaims{
		"exp":            time.Now().Add(60 * time.Second).Unix(),
		"sourceCrypto":   source,
		"targetCrypto":   target,
		"sourceAmount":   sourceAmount,
		"targetAmount":   targetAmount,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("error generating token: %v", err)
	}

	return tokenString, nil
}

func (s *BalanceService) GetExchangePreview(sourceCrypto, targetCrypto string, amount float64) (float64, string, error) {
	conversionRate, err := getCryptoPriceFromThirdParty(sourceCrypto, targetCrypto)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get price for %s: %v", sourceCrypto, err)
	}
	convertedAmount := amount * conversionRate

	token, err := createJWTToken(sourceCrypto, targetCrypto, amount, convertedAmount)
	if err != nil {
		return 0, "", fmt.Errorf("error in create JWT Toekn %s", err)
	}

	return convertedAmount, token, nil
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