package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/farisdewantoro/golang-day-trading-signal/models"
	"github.com/joho/godotenv"
)

// LoadConfig loads configuration from environment variables
func LoadConfig() *models.Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get stock symbols from environment
	stockSymbolsStr := getEnv("STOCK_SYMBOLS", "")
	var stockSymbols []string
	if stockSymbolsStr != "" {
		// Split comma-separated stock symbols
		for _, symbol := range strings.Split(stockSymbolsStr, ",") {
			symbol = strings.TrimSpace(symbol)
			if symbol != "" {
				stockSymbols = append(stockSymbols, symbol)
			}
		}
	}

	// If no stock symbols provided, use default
	if len(stockSymbols) == 0 {
		defaultSymbol := getEnv("DEFAULT_STOCK_SYMBOL", "INDY.JK")
		stockSymbols = []string{defaultSymbol}
	}

	config := &models.Config{
		GeminiAPIKey:       getEnv("GEMINI_API_KEY", ""),
		TelegramBotToken:   getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:     getEnv("TELEGRAM_CHAT_ID", ""),
		Port:               getEnv("PORT", "8080"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		DefaultStockSymbol: getEnv("DEFAULT_STOCK_SYMBOL", "INDY.JK"),
		SignalCooldownMins: getEnvAsInt("SIGNAL_COOLDOWN_MINUTES", 15),
		MinConfidenceLevel: getEnvAsInt("MIN_CONFIDENCE_LEVEL", 70),
		StockSymbols:       stockSymbols,
		WebhookURL:         getEnv("WEBHOOK_URL", ""),
	}

	log.Println(config)
	// Validate required configuration
	if config.GeminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY is required")
	}
	if config.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}
	if config.TelegramChatID == "" {
		log.Fatal("TELEGRAM_CHAT_ID is required")
	}

	return config
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Printf("Environment variable %s not found, using default value: %s", key, defaultValue)
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Environment variable %s has invalid integer value: %s, using default: %d", key, value, defaultValue)
	}
	return defaultValue
}
