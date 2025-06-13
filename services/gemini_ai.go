package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/farisdewantoro/golang-day-trading-signal/models"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiAIService handles AI-powered trading signal generation
type GeminiAIService struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiAIService creates a new Gemini AI service
func NewGeminiAIService(apiKey string) (*GeminiAIService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.0-flash")
	model.SetTemperature(0.7)
	model.SetTopP(0.8)
	model.SetTopK(40)

	return &GeminiAIService{
		client: client,
		model:  model,
	}, nil
}

// GenerateTradingSignal generates a trading signal using Gemini AI
func (g *GeminiAIService) GenerateTradingSignal(symbol string, ohlcData []models.OHLCData) (*models.TradingSignal, error) {
	prompt := g.buildPrompt(symbol, ohlcData)

	ctx := context.Background()
	resp, err := g.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response generated from Gemini")
	}

	content := resp.Candidates[0].Content
	if len(content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts in response")
	}

	text, _ := content.Parts[0].(genai.Text)
	signal, err := g.parseSignalResponse(string(text))
	if err != nil {
		return nil, fmt.Errorf("failed to parse signal response: %w", err)
	}

	signal.StockSymbol = symbol
	signal.GeneratedAt = time.Now()

	return signal, nil
}

// buildPrompt creates the prompt for Gemini AI
func (g *GeminiAIService) buildPrompt(symbol string, ohlcData []models.OHLCData) string {
	var dataBuilder strings.Builder
	dataBuilder.WriteString(fmt.Sprintf("Saya ingin kamu menganalisa saham %s yang diperdagangkan di Bursa Efek Indonesia. Data di bawah ini adalah candlestick 5-menit dengan range 1 hari sampai sekarang:\n\n", symbol))

	dataBuilder.WriteString("candlestick_data = [\n")
	for _, data := range ohlcData {
		dataBuilder.WriteString(fmt.Sprintf("{\n  \"timestamp\": \"%s\",\n  \"open\": %.2f,\n  \"high\": %.2f,\n  \"low\": %.2f,\n  \"close\": %.2f,\n  \"volume\": %d\n},\n",
			data.Timestamp.Format("2006-01-02T15:04:05-07:00"),
			data.Open, data.High, data.Low, data.Close, data.Volume))
	}
	dataBuilder.WriteString("]\n\n")

	prompt := fmt.Sprintf(`%s

### Instruksi:
Lakukan analisa teknikal berdasarkan data candlestick yang diberikan.

**PENTING: Risk-Reward Ratio 1:2**
- Setiap sinyal trading HARUS memiliki risk-reward ratio minimal 1:2
- Jika sinyal = "BUY": Target Price harus minimal 2x jarak dari Buy Price ke Stop Loss
- Jika sinyal = "SELL": Target Price harus minimal 2x jarak dari Sell Price ke Stop Loss
- Contoh: Buy Price = 1000, Stop Loss = 950 (risk = 50), maka Target Price minimal = 1100 (reward = 100, ratio = 1:2)
- Boleh dilonggarkan: Jika sinyal teknikal sangat kuat, rasio boleh sedikit di bawah 1:2
- Jika sinyal diberikan meskipun rasio < 1:2, jelaskan alasan validitas sinyal dengan jelas

Berikan sinyal trading:
- Sinyal: "BUY", "SELL", atau "WAIT"
- Harga Beli Ideal
- Target Jual (harus memenuhi risk-reward ratio 1:2)
- Stop Loss
- Confidence Level (0-100%%)
- Gunakan analisis teknikal untuk mendeteksi:
	- Pola candlestick seperti bullish engulfing, doji, hammer
	- Crossover antara 5EMA dan 20EMA
	- Level support dan resistance berdasarkan 2 jam terakhir
	- RSI dan MACD
	- Breakout harga dengan volume tinggi
- Jelaskan alasan di balik sinyal tersebut (berdasarkan analisa teknikal)

**Tambahan: Analisa OHLCV Terkini**
- Open: Harga pembukaan sesi 1/2
- High: Harga tertinggi terakhir
- Low: Harga terendah terakhir
- Close: Harga penutupan terakhir
- Volume: Volume perdagangan terakhir
- Penjelasan OHLCV:
  Tulis penjelasan dalam format naratif. Contoh:
  "Harga pembukaan sesi pertama berada di [OpenSesi1], sementara sesi kedua dibuka di [OpenSesi2]. Sepanjang hari, harga mencapai titik tertinggi di [High] dan terendah di [Low]. Saham ditutup di harga [Close] dengan total volume perdagangan sebesar [Volume]. Pola pergerakan harga menunjukkan [...analisa teknikal seperti bullish/bearish/momentum volume...]."

Berikan output dalam format JSON:
{
  "signal": "BUY",
  "buy_price": 2750,
  "target_price": 2850,
  "stop_loss": 2725,
  "confidence": 82,
  "reason": "Terjadi pola bullish engulfing pada timeframe 5 menit. Risk-reward ratio 1:2 terpenuhi (risk: 25, reward: 100).",
  "ohlcv_analysis": {
    "open": 2740,
    "high": 2750,
    "low": 2730,
    "close": 2745,
    "volume": 80000,
    "explanation": "Harga pembukaan sesi pertama berada di [OpenSesi1], sementara sesi kedua dibuka di [OpenSesi2]. Sepanjang hari, harga mencapai titik tertinggi di [High] dan terendah di [Low]. Saham ditutup di harga [Close] dengan total volume perdagangan sebesar [Volume]. Pola pergerakan harga menunjukkan [...analisa teknikal seperti bullish/bearish/momentum volume...]."
  }
}`, dataBuilder.String())

	return prompt
}

// parseSignalResponse parses the JSON response from Gemini
func (g *GeminiAIService) parseSignalResponse(text string) (*models.TradingSignal, error) {
	// Extract JSON from the response (in case there's extra text)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")

	if start == -1 || end == -1 {
		return nil, fmt.Errorf("no JSON found in response")
	}

	jsonStr := text[start : end+1]

	var signal models.TradingSignal
	if err := json.Unmarshal([]byte(jsonStr), &signal); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signal JSON: %w", err)
	}

	return &signal, nil
}

// Close closes the Gemini client
func (g *GeminiAIService) Close() error {
	return g.client.Close()
}

var buyDataExample = []models.OHLCData{
	{
		Timestamp: time.Date(2025, 6, 12, 9, 0, 0, 0, time.Local),
		Open:      1000,
		High:      1015,
		Low:       995,
		Close:     1010, // green candle
		Volume:    100000,
	},
	{
		Timestamp: time.Date(2025, 6, 12, 9, 5, 0, 0, time.Local),
		Open:      1010,
		High:      1020,
		Low:       1005,
		Close:     1015, // green candle
		Volume:    120000,
	},
	{
		Timestamp: time.Date(2025, 6, 12, 9, 10, 0, 0, time.Local),
		Open:      1015,
		High:      1030,
		Low:       1010,
		Close:     1025, // green candle
		Volume:    150000,
	},
	{
		Timestamp: time.Date(2025, 6, 12, 9, 15, 0, 0, time.Local),
		Open:      1025,
		High:      1040,
		Low:       1020,
		Close:     1035,   // breakout candle
		Volume:    180000, // volume tinggi mendukung breakout
	},
	{
		Timestamp: time.Date(2025, 6, 12, 9, 20, 0, 0, time.Local),
		Open:      1035,
		High:      1050,
		Low:       1030,
		Close:     1045,
		Volume:    200000, // volume terus meningkat
	},
}
