package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/farisdewantoro/golang-day-trading-signal/config"
	"github.com/farisdewantoro/golang-day-trading-signal/handlers"
	"github.com/farisdewantoro/golang-day-trading-signal/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting Trading Signal Service on port %s", cfg.Port)

	// Create trading signal service
	tradingService, err := services.NewTradingSignalService(cfg)
	if err != nil {
		log.Fatalf("Failed to create trading service: %v", err)
	}
	defer tradingService.Close()

	// Create and start cron scheduler if schedule times are configured
	var cronScheduler *services.CronScheduler
	if len(cfg.CronScheduleTimes) > 0 {
		cronScheduler, err = services.NewCronScheduler(tradingService, cfg.CronScheduleTimes)
		if err != nil {
			log.Printf("Failed to create cron scheduler: %v", err)
		} else {
			if err := cronScheduler.Start(); err != nil {
				log.Printf("Failed to start cron scheduler: %v", err)
			} else {
				log.Printf("Cron scheduler started with %d schedule times", len(cfg.CronScheduleTimes))
			}
		}
	} else {
		log.Println("No cron schedule times configured, skipping cron scheduler")
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Create handler
	signalHandler := handlers.NewSignalHandler(tradingService)

	// Set cron scheduler to handler if available
	if cronScheduler != nil {
		signalHandler.SetCronScheduler(cronScheduler)
	}

	// Setup routes
	api := router.Group("/api/v1")
	{
		api.GET("/health", signalHandler.HealthCheck)
		api.GET("/signal", signalHandler.GetSignal)
		api.POST("/signal", signalHandler.GenerateSignal)
		api.GET("/signal-all", signalHandler.GetSignalAll)
		api.GET("/signal-all-summary", signalHandler.GetSignalAllSummary)
		api.GET("/cron-status", signalHandler.GetCronStatus)
		api.POST("/webhook/setup", signalHandler.SetupWebhook)
		api.DELETE("/webhook", signalHandler.DeleteWebhook)
	}

	// Setup Telegram webhook route
	router.POST("/webhook/telegram", signalHandler.TelegramWebhook)

	// Setup webhook if in production
	if cfg.Environment == "production" && cfg.WebhookURL != "" {
		telegramService := tradingService.GetTelegramService()

		if err := telegramService.SetupWebhook(cfg.WebhookURL); err != nil {
			log.Printf("Failed to setup webhook: %v", err)
		} else {
			log.Printf("Telegram webhook setup successfully: %s", cfg.WebhookURL)
		}
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started successfully on port %s", cfg.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Stop cron scheduler if running
	if cronScheduler != nil {
		cronScheduler.Stop()
		log.Println("Cron scheduler stopped")
	}

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
