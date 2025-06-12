package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/farisdewantoro/golang-day-trading-signal/models"
	"github.com/farisdewantoro/golang-day-trading-signal/services"
	"github.com/gin-gonic/gin"
)

// SignalHandler handles HTTP requests for trading signals
type SignalHandler struct {
	tradingService *services.TradingSignalService
	cronScheduler  *services.CronScheduler
}

// NewSignalHandler creates a new signal handler
func NewSignalHandler(tradingService *services.TradingSignalService) *SignalHandler {
	return &SignalHandler{
		tradingService: tradingService,
	}
}

// SetCronScheduler sets the cron scheduler for the handler
func (h *SignalHandler) SetCronScheduler(cronScheduler *services.CronScheduler) {
	h.cronScheduler = cronScheduler
}

// GenerateSignal handles POST requests to generate trading signals
func (h *SignalHandler) GenerateSignal(c *gin.Context) {
	var req models.SignalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Use default symbol if not provided
	if req.StockSymbol == "" {
		req.StockSymbol = "INDY.JK"
	}

	// Generate signal
	signal, err := h.tradingService.GenerateSignal(req.StockSymbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trading signal generated successfully",
		Data:    signal,
	})
}

// GetSignal handles GET requests to generate trading signals
func (h *SignalHandler) GetSignal(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		symbol = "INDY.JK"
	}

	// Generate signal
	signal, err := h.tradingService.GenerateSignal(symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trading signal generated successfully",
		Data:    signal,
	})
}

// HealthCheck handles health check requests
func (h *SignalHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Trading Signal Service is healthy",
	})
}

// GetSignalAll handles GET requests to generate signals for all configured stocks
func (h *SignalHandler) GetSignalAll(c *gin.Context) {
	// Start bulk signal analysis in background
	h.tradingService.GenerateAllSignals()

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Bulk signal analysis started. Signals will be sent to Telegram as they are generated.",
		Data: map[string]interface{}{
			"status":  "started",
			"message": "Analysis is running in background. Check Telegram for results.",
		},
	})
}

// GetSignalAllSummary handles GET requests to generate signals for all configured stocks (summary only)
func (h *SignalHandler) GetSignalAllSummary(c *gin.Context) {
	// Start bulk signal analysis in background (summary only)
	h.tradingService.GenerateAllSignalsSummary()

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Bulk signal analysis started. You will receive a summary on Telegram once complete.",
		Data: map[string]interface{}{
			"status":  "started",
			"message": "Analysis is running in background. You will receive a 'request received' message and then a summary on Telegram.",
		},
	})
}

// TelegramWebhook handles incoming webhook messages from Telegram
func (h *SignalHandler) TelegramWebhook(c *gin.Context) {
	var webhook models.TelegramWebhook
	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid webhook format",
		})
		return
	}

	// Check if message exists
	if webhook.Message == nil {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "No message in webhook",
		})
		return
	}

	// Convert chat ID to string
	chatID := fmt.Sprintf("%d", webhook.Message.Chat.ID)
	text := strings.TrimSpace(webhook.Message.Text)

	// Get telegram service
	telegramService := h.tradingService.GetTelegramService()

	// Handle different commands and messages
	switch {
	case text == "/start":
		err := telegramService.SendWelcomeMessage(chatID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   "Failed to send welcome message",
			})
			return
		}

	case text == "/help":
		err := telegramService.SendHelpMessage(chatID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   "Failed to send help message",
			})
			return
		}

	case text == "/bulk":
		// Start bulk analysis
		go h.tradingService.GenerateAllSignals()
		err := telegramService.SendMessageToChat(chatID, "🚀 Starting bulk analysis for all configured stocks. You will receive signals as they are generated.")
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   "Failed to send bulk analysis message",
			})
			return
		}

	case text == "/summary":
		// Start bulk analysis with summary
		go h.tradingService.GenerateAllSignalsSummary()
		err := telegramService.SendMessageToChat(chatID, "📊 Starting bulk analysis with summary. You will receive a summary once complete.")
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   "Failed to send summary message",
			})
			return
		}

	case text == "/stocks":
		// Show configured stocks list
		stockSymbols := h.tradingService.GetConfiguredStocks()
		err := telegramService.SendStocksListMessage(chatID, stockSymbols)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   "Failed to send stocks list message",
			})
			return
		}

	case text != "" && !strings.HasPrefix(text, "/"):
		// Treat as stock symbol
		go h.handleStockSymbolRequest(chatID, text)

	default:
		// Unknown command
		err := telegramService.SendMessageToChat(chatID, "❓ Unknown command. Send /help for available commands.")
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.APIResponse{
				Success: false,
				Error:   "Failed to send unknown command message",
			})
			return
		}
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Webhook processed successfully",
	})
}

// handleStockSymbolRequest handles individual stock symbol requests
func (h *SignalHandler) handleStockSymbolRequest(chatID, symbol string) {
	telegramService := h.tradingService.GetTelegramService()

	// Send processing message
	processingMsg := fmt.Sprintf("🔍 Analyzing %s... Please wait.", symbol)
	telegramService.SendMessageToChat(chatID, processingMsg)

	// Generate signal
	signal, err := h.tradingService.GenerateSignal(symbol)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Failed to analyze %s: %s", symbol, err.Error())
		telegramService.SendMessageToChat(chatID, errorMsg)
		return
	}

	// Send signal to user
	err = telegramService.SendTradingSignalToChat(chatID, signal)
	if err != nil {
		errorMsg := fmt.Sprintf("❌ Failed to send signal for %s: %s", symbol, err.Error())
		telegramService.SendMessageToChat(chatID, errorMsg)
		return
	}
}

// SetupWebhook handles webhook setup requests
func (h *SignalHandler) SetupWebhook(c *gin.Context) {
	var req struct {
		WebhookURL string `json:"webhook_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	telegramService := h.tradingService.GetTelegramService()

	if err := telegramService.SetupWebhook(req.WebhookURL); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to setup webhook: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Webhook setup successfully",
		Data: map[string]interface{}{
			"webhook_url": req.WebhookURL,
		},
	})
}

// DeleteWebhook handles webhook deletion requests
func (h *SignalHandler) DeleteWebhook(c *gin.Context) {
	telegramService := h.tradingService.GetTelegramService()

	if err := telegramService.DeleteWebhook(); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to delete webhook: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Webhook deleted successfully",
	})
}

// GetCronStatus handles GET requests to get cron scheduler status
func (h *SignalHandler) GetCronStatus(c *gin.Context) {
	if h.cronScheduler == nil {
		c.JSON(http.StatusOK, models.APIResponse{
			Success: true,
			Message: "Cron scheduler is not configured",
			Data: map[string]interface{}{
				"enabled": false,
				"message": "No cron schedule times configured",
			},
		})
		return
	}

	scheduleInfo := h.cronScheduler.GetScheduleInfo()

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Cron scheduler status retrieved successfully",
		Data:    scheduleInfo,
	})
}
