package services

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

// CronScheduler handles scheduled execution of trading signals
type CronScheduler struct {
	cron           *cron.Cron
	tradingService *TradingSignalService
	scheduleTimes  []string
	timezone       *time.Location
}

// NewCronScheduler creates a new cron scheduler
func NewCronScheduler(tradingService *TradingSignalService, scheduleTimes []string) (*CronScheduler, error) {
	// Set timezone to WIB (UTC+7)
	wib, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("Failed to load WIB timezone, using UTC: %v", err)
		wib = time.UTC
	}

	scheduler := &CronScheduler{
		cron:           cron.New(cron.WithLocation(wib)),
		tradingService: tradingService,
		scheduleTimes:  scheduleTimes,
		timezone:       wib,
	}

	return scheduler, nil
}

// Start starts the cron scheduler
func (cs *CronScheduler) Start() error {
	log.Printf("Starting cron scheduler with WIB timezone")
	log.Printf("Configured execution times: %v", cs.scheduleTimes)

	// Add cron jobs for each schedule time
	for _, scheduleTime := range cs.scheduleTimes {
		scheduleTime = strings.TrimSpace(scheduleTime)
		if scheduleTime == "" {
			continue
		}

		// Parse time in format "HH:MM" (e.g., "08:30", "12:00", "14:45")
		parts := strings.Split(scheduleTime, ":")
		if len(parts) != 2 {
			log.Printf("Invalid time format: %s, expected HH:MM", scheduleTime)
			continue
		}

		hour := parts[0]
		minute := parts[1]

		// Create cron expression: "minute hour * * *" (every day at specified time)
		cronExpr := fmt.Sprintf("%s %s * * *", minute, hour)

		log.Printf("Adding cron job: %s (executes daily at %s WIB)", cronExpr, scheduleTime)

		_, err := cs.cron.AddFunc(cronExpr, func() {
			cs.executeTradingSignal()
		})

		if err != nil {
			log.Printf("Failed to add cron job for %s: %v", scheduleTime, err)
			continue
		}
	}

	// Start the cron scheduler
	cs.cron.Start()

	log.Printf("Cron scheduler started successfully with %d jobs", len(cs.cron.Entries()))
	return nil
}

// Stop stops the cron scheduler
func (cs *CronScheduler) Stop() {
	log.Printf("Stopping cron scheduler")
	cs.cron.Stop()
}

// executeTradingSignal executes the trading signal generation
func (cs *CronScheduler) executeTradingSignal() {
	now := time.Now().In(cs.timezone)
	log.Printf("🕐 [CRON] Executing scheduled trading signal generation at %s WIB", now.Format("2006-01-02 15:04:05"))

	// Execute the trading signal generation
	cs.tradingService.GenerateAllSignalsSummary()

	log.Printf("✅ [CRON] Scheduled trading signal generation completed at %s WIB", time.Now().In(cs.timezone).Format("2006-01-02 15:04:05"))
}

// GetNextRuns returns the next scheduled execution times
func (cs *CronScheduler) GetNextRuns() []time.Time {
	var nextRuns []time.Time

	for _, entry := range cs.cron.Entries() {
		nextRuns = append(nextRuns, entry.Next)
	}

	return nextRuns
}

// GetScheduleInfo returns information about the current schedule
func (cs *CronScheduler) GetScheduleInfo() map[string]interface{} {
	nextRuns := cs.GetNextRuns()

	info := map[string]interface{}{
		"timezone":         cs.timezone.String(),
		"configured_times": cs.scheduleTimes,
		"active_jobs":      len(cs.cron.Entries()),
		"next_runs":        nextRuns,
	}

	return info
}
