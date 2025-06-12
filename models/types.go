package models

import "time"

// OHLCData represents a single candlestick data point
type OHLCData struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
}

// OHLCVAnalysis represents the latest OHLCV data with explanation
type OHLCVAnalysis struct {
	Open        float64 `json:"open"`
	High        float64 `json:"high"`
	Low         float64 `json:"low"`
	Close       float64 `json:"close"`
	Volume      int64   `json:"volume"`
	Explanation string  `json:"explanation"`
}

// TradingSignal represents the AI-generated trading signal
type TradingSignal struct {
	Signal        string         `json:"signal"` // "BUY", "SELL", or "WAIT"
	BuyPrice      float64        `json:"buy_price"`
	TargetPrice   float64        `json:"target_price"`
	StopLoss      float64        `json:"stop_loss"`
	Confidence    int            `json:"confidence"` // 0-100
	NewsSummary   string         `json:"news_summary"`
	Reason        string         `json:"reason"`
	StockSymbol   string         `json:"stock_symbol"`
	GeneratedAt   time.Time      `json:"generated_at"`
	OHLCVAnalysis *OHLCVAnalysis `json:"ohlcv_analysis,omitempty"`
}

// YahooFinanceResponse represents the response from Yahoo Finance API
type YahooFinanceResponse struct {
	Chart struct {
		Result []struct {
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SignalRequest represents a request to generate a trading signal
type SignalRequest struct {
	StockSymbol string `json:"stock_symbol"`
}

// Config represents application configuration
type Config struct {
	GeminiAPIKey       string
	TelegramBotToken   string
	TelegramChatID     string
	Port               string
	Environment        string
	DefaultStockSymbol string
	SignalCooldownMins int
	MinConfidenceLevel int
	StockSymbols       []string // List of stock symbols to analyze
	WebhookURL         string   // Telegram webhook URL
}

// SignalSummary represents a summary of all analyzed signals
type SignalSummary struct {
	TotalAnalyzed int              `json:"total_analyzed"`
	BuySignals    []*TradingSignal `json:"buy_signals"`
	SellSignals   []*TradingSignal `json:"sell_signals"`
	HoldSignals   []*TradingSignal `json:"hold_signals"`
	FailedSignals []string         `json:"failed_signals"`
	GeneratedAt   time.Time        `json:"generated_at"`
}

// BulkSignalResult represents the result of bulk signal analysis
type BulkSignalResult struct {
	JobID       string         `json:"job_id"`
	Status      string         `json:"status"` // "started", "completed", "failed"
	Message     string         `json:"message"`
	Summary     *SignalSummary `json:"summary,omitempty"`
	StartedAt   time.Time      `json:"started_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// TelegramWebhook represents incoming webhook from Telegram
type TelegramWebhook struct {
	UpdateID int64            `json:"update_id"`
	Message  *TelegramMessage `json:"message,omitempty"`
}

// TelegramMessage represents a message from Telegram
type TelegramMessage struct {
	MessageID int64         `json:"message_id"`
	From      *TelegramUser `json:"from"`
	Chat      *TelegramChat `json:"chat"`
	Date      int64         `json:"date"`
	Text      string        `json:"text,omitempty"`
}

// TelegramUser represents a Telegram user
type TelegramUser struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}

// TelegramChat represents a Telegram chat
type TelegramChat struct {
	ID    int64  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title,omitempty"`
}

// TelegramWebhookResponse represents response to Telegram webhook
type TelegramWebhookResponse struct {
	Method      string               `json:"method"`
	ChatID      string               `json:"chat_id"`
	Text        string               `json:"text,omitempty"`
	ParseMode   string               `json:"parse_mode,omitempty"`
	ReplyMarkup *TelegramReplyMarkup `json:"reply_markup,omitempty"`
}

// TelegramReplyMarkup represents inline keyboard markup
type TelegramReplyMarkup struct {
	InlineKeyboard [][]TelegramInlineKeyboardButton `json:"inline_keyboard,omitempty"`
}

// TelegramInlineKeyboardButton represents an inline keyboard button
type TelegramInlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}
