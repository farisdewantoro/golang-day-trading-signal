package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"log"

	"github.com/farisdewantoro/golang-day-trading-signal/models"
)

// TelegramService handles sending messages via Telegram bot
type TelegramService struct {
	botToken string
	chatID   string
	client   *http.Client
}

// NewTelegramService creates a new Telegram service
func NewTelegramService(botToken, chatID string) *TelegramService {
	return &TelegramService{
		botToken: botToken,
		chatID:   chatID,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// SendTradingSignal sends a trading signal to Telegram
func (t *TelegramService) SendTradingSignal(signal *models.TradingSignal) error {
	message := t.formatSignalMessage(signal)
	return t.sendMessage(message)
}

// SendSignalSummary sends a summary of all analyzed signals to Telegram
func (t *TelegramService) SendSignalSummary(summary *models.SignalSummary) error {
	message := t.formatSummaryMessage(summary)
	return t.sendMessage(message)
}

// SendRequestReceivedMessage sends a message indicating that a bulk analysis request has been received
func (t *TelegramService) SendRequestReceivedMessage(totalStocks int) error {
	message := fmt.Sprintf(`📋 <b>BULK ANALYSIS REQUEST RECEIVED</b> 📋

📊 <b>Analysis Details:</b>
   📈 Total Stocks: %d
   ⏱️ Estimated Time: %d minutes
   🔄 Status: Processing...

⏰ <b>Request Time:</b> %s

Please wait while we analyze all stocks. You will receive a summary once the analysis is complete.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		totalStocks,
		(totalStocks*3)/60+1, // 3 seconds per stock + 1 minute buffer
		time.Now().Format("2006-01-02 15:04:05"))

	return t.sendMessage(message)
}

// SetupWebhook sets up the Telegram webhook URL
func (t *TelegramService) SetupWebhook(webhookURL string) error {
	baseURL := "https://api.telegram.org/bot" + t.botToken + "/setWebhook"

	params := url.Values{}
	params.Add("url", webhookURL)

	req, err := http.NewRequest("POST", baseURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create webhook setup request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to setup webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteWebhook removes the current webhook
func (t *TelegramService) DeleteWebhook() error {
	baseURL := "https://api.telegram.org/bot" + t.botToken + "/deleteWebhook"

	req, err := http.NewRequest("POST", baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create webhook delete request: %w", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendWelcomeMessage sends a welcome message with instructions
func (t *TelegramService) SendWelcomeMessage(chatID string) error {
	message := `🤖 <b>Welcome to Trading Signal Bot!</b> 🤖

📈 <b>Available Commands:</b>

🔍 <b>Single Stock Analysis:</b>
   Send a stock symbol (e.g., BBCA, BBRI, ANTM)
   Example: <code>ANTM</code>

📊 <b>Bulk Analysis:</b>
   /bulk - Analyze all configured stocks
   /summary - Get summary of all stocks
   /stocks - Show all configured stocks

❓ <b>Help:</b>
   /help - Show this help message
   /start - Start the bot


━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	return t.sendMessageToChat(chatID, message)
}

// SendHelpMessage sends a help message
func (t *TelegramService) SendHelpMessage(chatID string) error {
	message := `📚 <b>Help & Instructions</b> 📚

🔍 <b>How to use:</b>
   1. Send a stock symbol to get trading signal
   2. Wait for analysis to complete
   3. Receive detailed signal with buy/sell recommendations

📊 <b>Available Commands:</b>
   /stocks - Show all configured stocks
   /bulk - Analyze all configured stocks (individual signals)
   /summary - Analyze all configured stocks (summary only)
   /help - Show this help message
   /start - Start the bot

📊 <b>Signal Types:</b>
   🟢 BUY - Good opportunity to buy
   🔴 SELL - Consider selling
   🟡 WAIT - Hold current position

💰 <b>Signal Information:</b>
   • Buy Price: Recommended entry price
   • Target Price: Profit target
   • Stop Loss: Risk management level
   • Confidence: AI confidence level (0-100%)
   • Risk-Reward Ratio: Risk vs potential reward

⚠️ <b>Disclaimer:</b>
   This is for educational purposes only.
   Always do your own research before trading.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`

	return t.sendMessageToChat(chatID, message)
}

// sendMessage sends a text message to Telegram
func (t *TelegramService) sendMessage(message string) error {
	baseURL := "https://api.telegram.org/bot" + t.botToken + "/sendMessage"

	params := url.Values{}
	params.Add("chat_id", t.chatID)
	params.Add("text", message)
	params.Add("parse_mode", "HTML")

	req, err := http.NewRequest("POST", baseURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// sendMessageToChat sends a message to a specific chat ID
func (t *TelegramService) sendMessageToChat(chatID, message string) error {
	baseURL := "https://api.telegram.org/bot" + t.botToken + "/sendMessage"

	params := url.Values{}
	params.Add("chat_id", chatID)
	params.Add("text", message)
	params.Add("parse_mode", "HTML")

	req, err := http.NewRequest("POST", baseURL, strings.NewReader(params.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API returned status: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendMessageToChat sends a message to a specific chat ID (public method)
func (t *TelegramService) SendMessageToChat(chatID, message string) error {
	return t.sendMessageToChat(chatID, message)
}

// SendTradingSignalToChat sends a trading signal to a specific chat ID
func (t *TelegramService) SendTradingSignalToChat(chatID string, signal *models.TradingSignal) error {
	message := t.formatSignalMessage(signal)
	return t.sendMessageToChat(chatID, message)
}

// calculateRiskRewardRatio calculates the risk-reward ratio for a trading signal
func (t *TelegramService) calculateRiskRewardRatio(signal *models.TradingSignal) (float64, float64, float64, error) {
	if signal.Signal == "WAIT" {
		return 0, 0, 0, nil
	}

	var risk, reward float64

	if signal.Signal == "BUY" {
		risk = signal.BuyPrice - signal.StopLoss
		reward = signal.TargetPrice - signal.BuyPrice
	} else if signal.Signal == "SELL" {
		risk = signal.StopLoss - signal.BuyPrice
		reward = signal.BuyPrice - signal.TargetPrice
	} else {
		log.Println(fmt.Errorf("invalid signal type: %s", signal.Signal))
		return 0, 0, 0, nil
	}

	// Check if risk and reward are positive
	if risk <= 0 {
		log.Println(fmt.Errorf("invalid risk calculation: risk must be positive, got %.2f", risk))
		return 0, 0, 0, nil
	}
	if reward <= 0 {
		log.Println(fmt.Errorf("invalid reward calculation: reward must be positive, got %.2f", reward))
		return 0, 0, 0, nil
	}

	ratio := reward / risk
	return risk, reward, ratio, nil
}

// formatSignalMessage formats the trading signal for Telegram
func (t *TelegramService) formatSignalMessage(signal *models.TradingSignal) string {
	var emoji string
	switch strings.ToUpper(signal.Signal) {
	case "BUY":
		emoji = "🟢"
	case "SELL":
		emoji = "🔴"
	case "WAIT":
		emoji = "🟡"
	default:
		emoji = "⚪"
	}

	// Create title with emoji and signal type
	title := fmt.Sprintf("%s <b>TRADING SIGNAL: %s %s</b> %s",
		emoji,
		strings.ToUpper(signal.Signal),
		signal.StockSymbol,
		emoji)

	message := fmt.Sprintf(`%s

💰 <b>Buy Price:</b> $%.2f
🎯 <b>Target Price:</b> $%.2f
🛑 <b>Stop Loss:</b> $%.2f`,
		title,
		signal.BuyPrice,
		signal.TargetPrice,
		signal.StopLoss)

	// Add risk-reward ratio if not WAIT signal
	if signal.Signal != "WAIT" {
		risk, reward, ratio, err := t.calculateRiskRewardRatio(signal)
		if err == nil {
			message += fmt.Sprintf(`

⚖️ <b>Risk-Reward Analysis:</b>
   💸 Risk: $%.2f
   💰 Reward: $%.2f
   📊 Ratio: 1:%.2f`,
				risk, reward, ratio)
		}
	}

	message += fmt.Sprintf(`

📈 <b>Confidence Level:</b> %d%%

📝 <b>Signal Reason:</b>
%s`,
		signal.Confidence,
		signal.Reason)

	// Add OHLCV analysis if available
	if signal.OHLCVAnalysis != nil {
		message += fmt.Sprintf(`

📊 <b>Current OHLCV Data:</b>
   📈 Open: $%.2f
   🔺 High: $%.2f
   🔻 Low: $%.2f
   📉 Close: $%.2f
   📊 Volume: %d

📋 <b>Technical Analysis:</b>
%s`,
			signal.OHLCVAnalysis.Open,
			signal.OHLCVAnalysis.High,
			signal.OHLCVAnalysis.Low,
			signal.OHLCVAnalysis.Close,
			signal.OHLCVAnalysis.Volume,
			signal.OHLCVAnalysis.Explanation)
	}

	message += fmt.Sprintf(`

⏰ <b>Generated At:</b> %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		signal.GeneratedAt.Format("2006-01-02 15:04:05"))

	return message
}

// formatSummaryMessage formats the signal summary for Telegram
func (t *TelegramService) formatSummaryMessage(summary *models.SignalSummary) string {
	title := "📊 <b>BULK SIGNAL ANALYSIS SUMMARY</b> 📊"

	message := fmt.Sprintf(`%s

📈 <b>Analysis Results:</b>
   ✅ Total Analyzed: %d stocks
   🟢 Buy Signals: %d
   🔴 Sell Signals: %d
   🟡 Hold Signals: %d
   ❌ Failed: %d

⏰ <b>Generated At:</b> %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		title,
		summary.TotalAnalyzed,
		len(summary.BuySignals),
		len(summary.SellSignals),
		len(summary.HoldSignals),
		len(summary.FailedSignals),
		summary.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Add buy signals
	if len(summary.BuySignals) > 0 {
		message += "\n\n🟢 <b>BUY SIGNALS:</b>"
		for _, signal := range summary.BuySignals {
			_, _, ratio, err := t.calculateRiskRewardRatio(signal)
			if err == nil {
				message += fmt.Sprintf("\n   • %s - Confidence: %d%% - Buy: $%.2f - Target: $%.2f - Cut Loss: $%.2f - R:R 1:%.2f",
					signal.StockSymbol, signal.Confidence, signal.BuyPrice, signal.TargetPrice, signal.StopLoss, ratio)
			} else {
				message += fmt.Sprintf("\n   • %s - Confidence: %d%% - Buy: $%.2f - Target: $%.2f - Cut Loss: $%.2f",
					signal.StockSymbol, signal.Confidence, signal.BuyPrice, signal.TargetPrice, signal.StopLoss)
			}
		}
	}

	// Add sell signals
	if len(summary.SellSignals) > 0 {
		message += "\n\n🔴 <b>SELL SIGNALS:</b>"
		for _, signal := range summary.SellSignals {
			_, _, ratio, err := t.calculateRiskRewardRatio(signal)
			if err == nil {
				message += fmt.Sprintf("\n   • %s - Confidence: %d%% - Stop Loss: $%.2f - R:R 1:%.2f",
					signal.StockSymbol, signal.Confidence, signal.StopLoss, ratio)
			} else {
				message += fmt.Sprintf("\n   • %s - Confidence: %d%% - Stop Loss: $%.2f",
					signal.StockSymbol, signal.Confidence, signal.StopLoss)
			}
		}
	}

	// Add hold signals
	if len(summary.HoldSignals) > 0 {
		message += "\n\n🟡 <b>HOLD SIGNALS:</b>"
		for _, signal := range summary.HoldSignals {
			message += fmt.Sprintf("\n   • %s - Confidence: %d%%",
				signal.StockSymbol, signal.Confidence)
		}
	}

	// Add failed signals
	if len(summary.FailedSignals) > 0 {
		message += "\n\n❌ <b>FAILED ANALYSIS:</b>"
		for _, symbol := range summary.FailedSignals {
			message += fmt.Sprintf("\n   • %s", symbol)
		}
	}

	message += "\n\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

	return message
}

// SendStocksListMessage sends a message with all configured stock symbols
func (t *TelegramService) SendStocksListMessage(chatID string, stockSymbols []string) error {
	title := "📋 <b>CONFIGURED STOCKS LIST</b> 📋"

	message := fmt.Sprintf(`%s

📈 <b>Total Stocks:</b> %d

📊 <b>Stock Symbols:</b>`,
		title,
		len(stockSymbols))

	// Group stocks in rows of 5 for better readability
	for i := 0; i < len(stockSymbols); i += 5 {
		end := i + 5
		if end > len(stockSymbols) {
			end = len(stockSymbols)
		}

		row := stockSymbols[i:end]
		message += "\n   "
		for j, symbol := range row {
			if j > 0 {
				message += " • "
			}
			message += fmt.Sprintf("<code>%s</code>", symbol)
		}
	}

	message += fmt.Sprintf(`

💡 <b>Usage:</b>
   • Send any symbol above to get trading signal
   • Use /bulk to analyze all stocks
   • Use /summary for bulk analysis summary

⏰ <b>Last Updated:</b> %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`,
		time.Now().Format("2006-01-02 15:04:05"))

	return t.sendMessageToChat(chatID, message)
}
