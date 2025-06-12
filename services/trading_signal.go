package services

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/farisdewantoro/golang-day-trading-signal/models"
)

// TradingSignalService orchestrates the entire trading signal generation process
type TradingSignalService struct {
	yahooService    *YahooFinanceService
	geminiService   *GeminiAIService
	telegramService *TelegramService
	config          *models.Config
	signalCache     map[string]time.Time
	cacheMutex      sync.RWMutex
}

// NewTradingSignalService creates a new trading signal service
func NewTradingSignalService(config *models.Config) (*TradingSignalService, error) {
	geminiService, err := NewGeminiAIService(config.GeminiAPIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini service: %w", err)
	}

	return &TradingSignalService{
		yahooService:    NewYahooFinanceService(),
		geminiService:   geminiService,
		telegramService: NewTelegramService(config.TelegramBotToken, config.TelegramChatID),
		config:          config,
		signalCache:     make(map[string]time.Time),
	}, nil
}

// GenerateSignal generates a trading signal for a given stock symbol
func (t *TradingSignalService) GenerateSignal(symbol string) (*models.TradingSignal, error) {

	log.Printf("Generating trading signal for %s", symbol)

	// Fetch OHLC data
	ohlcData, err := t.yahooService.FetchOHLCData(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OHLC data: %w", err)
	}

	if len(ohlcData) == 0 {
		return nil, fmt.Errorf("no OHLC data available for %s", symbol)
	}

	log.Printf("Fetched %d OHLC data points for %s", len(ohlcData), symbol)

	// Generate AI signal
	signal, err := t.geminiService.GenerateTradingSignal(symbol, ohlcData)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI signal: %w", err)
	}

	// // Send to Telegram if confidence is high enough
	// if err := t.telegramService.SendTradingSignal(signal); err != nil {
	// 	log.Printf("Failed to send signal to Telegram: %v", err)
	// 	// Don't return error here, as the signal was generated successfully
	// } else {
	// 	log.Printf("Signal sent to Telegram for %s", symbol)
	// }

	return signal, nil
}

// canGenerateSignal checks if enough time has passed since the last signal
func (t *TradingSignalService) canGenerateSignal(symbol string) bool {
	t.cacheMutex.RLock()
	defer t.cacheMutex.RUnlock()

	lastSignal, exists := t.signalCache[symbol]
	if !exists {
		return true
	}

	cooldownDuration := time.Duration(t.config.SignalCooldownMins) * time.Minute
	return time.Since(lastSignal) >= cooldownDuration
}

// updateSignalCache updates the signal generation timestamp
func (t *TradingSignalService) updateSignalCache(symbol string) {
	t.cacheMutex.Lock()
	defer t.cacheMutex.Unlock()
	t.signalCache[symbol] = time.Now()
}

// Close closes the service and its dependencies
func (t *TradingSignalService) Close() error {
	if t.geminiService != nil {
		return t.geminiService.Close()
	}
	return nil
}

// GetTelegramService returns the telegram service for external use
func (t *TradingSignalService) GetTelegramService() *TelegramService {
	return t.telegramService
}

// SendTradingSignalToChat sends a trading signal to a specific chat ID
func (t *TradingSignalService) SendTradingSignalToChat(chatID string, signal *models.TradingSignal) error {
	message := t.telegramService.formatSignalMessage(signal)
	return t.telegramService.sendMessageToChat(chatID, message)
}

// GenerateAllSignals generates signals for all configured stock symbols
func (t *TradingSignalService) GenerateAllSignals() {
	go func() {
		log.Printf("Starting bulk signal analysis for %d stocks", len(t.config.StockSymbols))

		var buySignals []*models.TradingSignal
		var sellSignals []*models.TradingSignal
		var holdSignals []*models.TradingSignal
		var failedSignals []string

		// Analyze each stock sequentially with 3-second delay
		for i, symbol := range t.config.StockSymbols {
			log.Printf("Analyzing stock %d/%d: %s", i+1, len(t.config.StockSymbols), symbol)

			// Generate signal for current stock
			signal, err := t.GenerateSignal(symbol)
			if err != nil {
				log.Printf("Failed to generate signal for %s: %v", symbol, err)
				failedSignals = append(failedSignals, symbol)
				continue
			}

			// Categorize signal
			switch strings.ToUpper(signal.Signal) {
			case "BUY":
				buySignals = append(buySignals, signal)
			case "SELL":
				sellSignals = append(sellSignals, signal)
			case "WAIT", "HOLD":
				holdSignals = append(holdSignals, signal)
			default:
				holdSignals = append(holdSignals, signal)
			}

			// Wait 3 seconds before next analysis (except for the last one)
			if i < len(t.config.StockSymbols)-1 {
				time.Sleep(3 * time.Second)
			}
		}

		// Create summary
		summary := &models.SignalSummary{
			TotalAnalyzed: len(t.config.StockSymbols),
			BuySignals:    buySignals,
			SellSignals:   sellSignals,
			HoldSignals:   holdSignals,
			FailedSignals: failedSignals,
			GeneratedAt:   time.Now(),
		}

		// Send summary to Telegram
		if err := t.telegramService.SendSignalSummary(summary); err != nil {
			log.Printf("Failed to send signal summary to Telegram: %v", err)
		} else {
			log.Printf("Signal summary sent to Telegram")
		}

		log.Printf("Bulk signal analysis completed. Total: %d, Buy: %d, Sell: %d, Hold: %d, Failed: %d",
			summary.TotalAnalyzed, len(buySignals), len(sellSignals), len(holdSignals), len(failedSignals))
	}()
}

// GenerateAllSignalsSummary generates signals for all configured stock symbols but only sends summary to Telegram
func (t *TradingSignalService) GenerateAllSignalsSummary() {
	go func() {
		log.Printf("Starting bulk signal analysis for %d stocks (summary only)", len(t.config.StockSymbols))

		// Send initial "request received" message
		if err := t.telegramService.SendRequestReceivedMessage(len(t.config.StockSymbols)); err != nil {
			log.Printf("Failed to send request received message to Telegram: %v", err)
		} else {
			log.Printf("Request received message sent to Telegram")
		}

		var buySignals []*models.TradingSignal
		var sellSignals []*models.TradingSignal
		var holdSignals []*models.TradingSignal
		var failedSignals []string

		// Analyze each stock sequentially with 3-second delay
		for i, symbol := range t.config.StockSymbols {
			log.Printf("Analyzing stock %d/%d: %s", i+1, len(t.config.StockSymbols), symbol)

			// Fetch OHLC data
			ohlcData, err := t.yahooService.FetchOHLCData(symbol)
			if err != nil {
				log.Printf("Failed to fetch OHLC data for %s: %v", symbol, err)
				failedSignals = append(failedSignals, symbol)
				continue
			}

			if len(ohlcData) == 0 {
				log.Printf("No OHLC data available for %s", symbol)
				failedSignals = append(failedSignals, symbol)
				continue
			}

			// Generate AI signal
			signal, err := t.geminiService.GenerateTradingSignal(symbol, ohlcData)
			if err != nil {
				log.Printf("Failed to generate AI signal for %s: %v", symbol, err)
				failedSignals = append(failedSignals, symbol)
				continue
			}

			// Categorize signal
			switch strings.ToUpper(signal.Signal) {
			case "BUY":
				buySignals = append(buySignals, signal)
			case "SELL":
				sellSignals = append(sellSignals, signal)
			case "WAIT", "HOLD":
				holdSignals = append(holdSignals, signal)
			default:
				holdSignals = append(holdSignals, signal)
			}

			// Wait 3 seconds before next analysis (except for the last one)
			if i < len(t.config.StockSymbols)-1 {
				time.Sleep(3 * time.Second)
			}
		}

		// Create summary
		summary := &models.SignalSummary{
			TotalAnalyzed: len(t.config.StockSymbols),
			BuySignals:    buySignals,
			SellSignals:   sellSignals,
			HoldSignals:   holdSignals,
			FailedSignals: failedSignals,
			GeneratedAt:   time.Now(),
		}

		// Send summary to Telegram
		if err := t.telegramService.SendSignalSummary(summary); err != nil {
			log.Printf("Failed to send signal summary to Telegram: %v", err)
		} else {
			log.Printf("Signal summary sent to Telegram")
		}

		log.Printf("Bulk signal analysis completed. Total: %d, Buy: %d, Sell: %d, Hold: %d, Failed: %d",
			summary.TotalAnalyzed, len(buySignals), len(sellSignals), len(holdSignals), len(failedSignals))
	}()
}

// GetConfiguredStocks returns the list of configured stock symbols
func (t *TradingSignalService) GetConfiguredStocks() []string {
	return t.config.StockSymbols
}
