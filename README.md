# 📈 Day Trading Signal App for Indonesian Stocks

A Go-based application that generates AI-powered trading signals for Indonesian stocks using Google Gemini AI. The app fetches real-time OHLC data from Yahoo Finance, analyzes it with AI, and sends trading signals via Telegram.

## 🚀 Features

- **Real-time OHLC Data**: Fetches 5-minute candlestick data from Yahoo Finance
- **AI-Powered Analysis**: Uses Google Gemini AI for technical and sentiment analysis
- **Telegram Integration**: Sends formatted trading signals to Telegram
- **RESTful API**: HTTP endpoints for manual and automated signal generation
- **Cooldown Protection**: Prevents signal spam with configurable cooldown periods
- **Confidence Filtering**: Only sends signals above minimum confidence threshold

## 📋 Prerequisites

- Go 1.21 or higher
- Google Gemini API key
- Telegram Bot Token and Chat ID
- Optional: News API key for enhanced sentiment analysis

## 🛠️ Installation

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd golang-day-trading-signal
   ```

2. **Install dependencies**:
   ```bash
   go mod tidy
   ```

3. **Set up environment variables**:
   ```bash
   cp env.example .env
   ```
   
   Edit `.env` with your actual API keys and configuration:
   ```env
   GEMINI_API_KEY=your_gemini_api_key_here
   TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
   TELEGRAM_CHAT_ID=your_telegram_chat_id_here
   PORT=8080
   ENVIRONMENT=development
   DEFAULT_STOCK_SYMBOL=INDY.JK
   STOCK_SYMBOLS=INDY.JK,BBRI.JK,TLKM.JK,ASII.JK,ICBP.JK
   SIGNAL_COOLDOWN_MINUTES=15
   MIN_CONFIDENCE_LEVEL=70
   ```

## 🔧 Configuration

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `GEMINI_API_KEY` | Google Gemini API key | `AIzaSy...` |
| `TELEGRAM_BOT_TOKEN` | Telegram bot token | `123456789:ABC...` |
| `TELEGRAM_CHAT_ID` | Telegram chat/channel ID | `-1001234567890` |

### Optional Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `ENVIRONMENT` | Environment mode | `development` |
| `DEFAULT_STOCK_SYMBOL` | Default stock symbol | `INDY.JK` |
| `STOCK_SYMBOLS` | Comma-separated list of stock symbols for bulk analysis | `DEFAULT_STOCK_SYMBOL` |
| `SIGNAL_COOLDOWN_MINUTES` | Minutes between signals | `15` |
| `MIN_CONFIDENCE_LEVEL` | Minimum confidence % | `70` |
| `NEWS_API_KEY` | News API key (optional) | `` |

## 🚀 Running the Application

### Development Mode
```bash
go run main.go
```

### Production Mode
```bash
go build -o trading-signal
./trading-signal
```

The server will start on `http://localhost:8080` (or your configured port).

## 📡 API Endpoints

### Health Check
```http
GET /api/v1/health
```

### Generate Signal (GET)
```http
GET /api/v1/signal?symbol=INDY.JK
```

### Generate Signal (POST)
```http
POST /api/v1/signal
Content-Type: application/json

{
  "stock_symbol": "INDY.JK"
}
```

### Generate Signals for All Stocks
```http
GET /api/v1/signal-all
```

This endpoint analyzes all stocks configured in `STOCK_SYMBOLS` environment variable. The analysis runs in the background and sends individual signals to Telegram as they are generated. After all stocks are analyzed, a summary message is sent to Telegram.

**Response:**
```json
{
  "success": true,
  "message": "Bulk signal analysis started. Signals will be sent to Telegram as they are generated.",
  "data": {
    "status": "started",
    "message": "Analysis is running in background. Check Telegram for results."
  }
}
```

## 📊 Example API Response

```json
{
  "success": true,
  "message": "Trading signal generated successfully",
  "data": {
    "signal": "BUY",
    "buy_price": 2750,
    "target_price": 2800,
    "stop_loss": 2725,
    "confidence": 82,
    "news_summary": "Harga batubara global naik 2%. Sentimen pasar terhadap sektor energi positif.",
    "reason": "Terjadi pola bullish engulfing pada timeframe 5 menit. Sentimen positif karena harga batubara global naik 2%.",
    "stock_symbol": "INDY.JK",
    "generated_at": "2024-01-15T10:30:00Z"
  }
}
```

## 🤖 Telegram Integration

The app sends formatted trading signals to Telegram with the following format:

```
🟢 Signal: BUY $INDY.JK
💰 Buy: 2750.00
🎯 Target: 2800.00
🛑 Stop Loss: 2725.00
📈 Confidence: 82%
🗞️ News: Harga batubara global naik 2%. Sentimen pasar terhadap sektor energi positif.
📝 Reason: Terjadi pola bullish engulfing pada timeframe 5 menit. Sentimen positif karena harga batubara global naik 2%.
⏰ Generated: 2024-01-15 17:30:00
```

### Telegram Output Preview

![Telegram Output Preview](assets/telegram_overview.gif)

*Watch the GIF above to see how trading signals appear in Telegram with real-time updates and interactive commands.*

### Telegram Bot Commands

The bot supports the following commands when used via webhook:

- `/start` - Welcome message with instructions
- `/help` - Detailed help and usage instructions
- `/stocks` - Show all configured stocks list
- `/bulk` - Analyze all configured stocks (individual signals)
- `/summary` - Analyze all configured stocks (summary only)
- `AAPL` - Send any stock symbol to get trading signal

### Webhook Setup

To enable Telegram webhook functionality:

1. Set the `WEBHOOK_URL` environment variable to your server's webhook endpoint
2. Set `ENVIRONMENT=production` to auto-setup webhook on startup
3. Or manually setup webhook using the API:

```http
POST /api/v1/webhook/setup
Content-Type: application/json

{
  "webhook_url": "https://437a-114-10-45-103.ngrok-free.app"
}
```

curl -X POST "http://localhost:8080/api/v1/webhook/setup" \
  -H "Content-Type: application/json" \
  -d '{
    "webhook_url": "https://437a-114-10-45-103.ngrok-free.app"
  }'

### Webhook Management

```http
# Setup webhook
POST /api/v1/webhook/setup
Content-Type: application/json

{
  "webhook_url": "https://your-domain.com/webhook/telegram"
}

# Delete webhook
DELETE /api/v1/webhook
```

### Webhook Endpoint

```http
POST /webhook/telegram
Content-Type: application/json

{
  "update_id": 123456789,
  "message": {
    "message_id": 1,
    "from": {
      "id": 123456789,
      "is_bot": false,
      "first_name": "User"
    },
    "chat": {
      "id": 123456789,
      "type": "private"
    },
    "date": 1642234567,
    "text": "AAPL"
  }
}
```

## 🔄 Automated Usage

### Cron Job Example
```bash
# Run every 15 minutes during market hours
*/15 9-15 * * 1-5 curl -X GET "http://localhost:8080/api/v1/signal?symbol=INDY.JK"
```

### Docker Example
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o trading-signal

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/trading-signal .
CMD ["./trading-signal"]
```

## 🏗️ Project Structure

```
├── main.go                 # Application entry point
├── go.mod                  # Go module file
├── env.example             # Environment variables template
├── README.md              # This file
├── config/
│   └── config.go          # Configuration management
├── models/
│   └── types.go           # Data structures and types
├── services/
│   ├── yahoo_finance.go   # Yahoo Finance API integration
│   ├── gemini_ai.go       # Google Gemini AI integration
│   ├── telegram.go        # Telegram bot integration
│   └── trading_signal.go  # Main trading signal service
└── handlers/
    └── signal_handler.go  # HTTP request handlers
```

## 🔒 Security Considerations

- Store API keys securely in environment variables
- Use HTTPS in production
- Implement rate limiting for API endpoints
- Monitor API usage and costs
- Validate all input data

## 🐛 Troubleshooting

### Common Issues

1. **"GEMINI_API_KEY is required"**
   - Ensure your `.env` file is properly configured
   - Check that the API key is valid and has sufficient quota

2. **"TELEGRAM_BOT_TOKEN is required"**
   - Create a Telegram bot via @BotFather
   - Get the chat ID where you want to receive signals

3. **"No data returned for symbol"**
   - Verify the stock symbol is correct (e.g., `INDY.JK` for Indonesian stocks)
   - Check if the market is open and data is available

4. **Import errors**
   - Run `go mod tidy` to download dependencies
   - Ensure Go version is 1.21 or higher

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📞 Support

For support and questions, please open an issue on GitHub or contact the maintainers. 