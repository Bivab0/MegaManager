package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Config holds all application configuration
type Config struct {
	// YouTube
	YouTubeAPIKey   string
	YouTubeChannelID string

	// Telegram
	TelegramBotToken string
	TelegramChatID   string

	// LLM
	LLMEndpoint string
	LLMAPIKey   string
	LLMModel    string

	// Intervals (in minutes)
	TrendScoutInterval   int
	PulseMonitorInterval int
	CommentPollInterval  int
}

// loadEnvFile loads .env file if it exists
func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return // .env file is optional
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Set environment variable if not already set
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	// Load .env file first
	loadEnvFile()

	cfg := &Config{
		YouTubeAPIKey:        os.Getenv("YOUTUBE_API_KEY"),
		YouTubeChannelID:     os.Getenv("YOUTUBE_CHANNEL_ID"),
		TelegramBotToken:     os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:       os.Getenv("TELEGRAM_CHAT_ID"),
		LLMEndpoint:          os.Getenv("LLM_API_ENDPOINT"),
		LLMAPIKey:            os.Getenv("LLM_API_KEY"),
		LLMModel:             getEnv("LLM_MODEL", "gpt-4o-mini"),
		TrendScoutInterval:   getEnvInt("TREND_SCOUT_INTERVAL", 720),
		PulseMonitorInterval: getEnvInt("PULSE_MONITOR_INTERVAL", 180),
		CommentPollInterval:  getEnvInt("COMMENT_POLL_INTERVAL", 5),
	}

	// Validate required fields
	if cfg.YouTubeAPIKey == "" {
		log.Fatal("YOUTUBE_API_KEY is required")
	}
	if cfg.YouTubeChannelID == "" {
		log.Fatal("YOUTUBE_CHANNEL_ID is required")
	}
	if cfg.TelegramBotToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.TelegramChatID == "" {
		log.Fatal("TELEGRAM_CHAT_ID is required")
	}
	if cfg.LLMEndpoint == "" {
		log.Fatal("LLM_API_ENDPOINT is required")
	}
	if cfg.LLMAPIKey == "" {
		log.Fatal("LLM_API_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func main() {
	log.Println("🚀 MegaManager AI starting...")

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize providers
	platform := NewYouTubePlatform(cfg.YouTubeAPIKey, cfg.YouTubeChannelID)
	ai := NewLLMClient(cfg.LLMEndpoint, cfg.LLMAPIKey, cfg.LLMModel)
	ui := NewTelegramUI(cfg.TelegramBotToken, cfg.TelegramChatID)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// WaitGroup to track all goroutines
	var wg sync.WaitGroup

	// Start Trend Scout (runs every N minutes)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runTrendScout(ctx, platform, ui, time.Duration(cfg.TrendScoutInterval)*time.Minute)
	}()

	// Start Pulse Monitor (runs every N minutes)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runPulseMonitor(ctx, platform, ui, time.Duration(cfg.PulseMonitorInterval)*time.Minute)
	}()

	// Start Smart Responder (polls every N minutes)
	wg.Add(1)
	go func() {
		defer wg.Done()
		runSmartResponder(ctx, platform, ai, ui, time.Duration(cfg.CommentPollInterval)*time.Minute)
	}()

	log.Println("✅ All services started successfully")
	log.Printf("📊 Trend Scout: every %d minutes", cfg.TrendScoutInterval)
	log.Printf("📈 Pulse Monitor: every %d minutes", cfg.PulseMonitorInterval)
	log.Printf("💬 Smart Responder: every %d minutes", cfg.CommentPollInterval)

	// Wait for shutdown signal
	<-sigChan
	log.Println("\n🛑 Shutdown signal received, stopping services...")

	// Cancel context to stop all goroutines
	cancel()

	// Wait for all goroutines to finish
	wg.Wait()

	log.Println("👋 MegaManager AI stopped gracefully")
}

// runTrendScout runs the trend scouting service on an interval
func runTrendScout(ctx context.Context, platform *YouTubePlatform, ui *TelegramUI, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	if err := trendScout(ctx, platform, ui); err != nil {
		log.Printf("❌ Trend Scout error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("🔥 Trend Scout stopping...")
			return
		case <-ticker.C:
			if err := trendScout(ctx, platform, ui); err != nil {
				log.Printf("❌ Trend Scout error: %v", err)
			}
		}
	}
}

// trendScout fetches trending videos and sends them to Telegram
func trendScout(ctx context.Context, platform *YouTubePlatform, ui *TelegramUI) error {
	log.Println("🔥 Running Trend Scout...")

	trending, err := platform.GetTrendingVideos(ctx, 5)
	if err != nil {
		return err
	}

	message := "🔥 Trending in Your Niche (Top 5):\n\n"
	if len(trending) == 0 {
		message += "No trending videos found in your category\n"
	} else {
		for i, video := range trending {
			message += formatVideo(i+1, video)
		}
	}
	message += "\n💡 Inspiration: Pick a topic and create your version!"

	return ui.SendMessage(ctx, message)
}

// runPulseMonitor runs the channel analytics service on an interval
func runPulseMonitor(ctx context.Context, platform *YouTubePlatform, ui *TelegramUI, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	if err := pulseMonitor(ctx, platform, ui); err != nil {
		log.Printf("❌ Pulse Monitor error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("📈 Pulse Monitor stopping...")
			return
		case <-ticker.C:
			if err := pulseMonitor(ctx, platform, ui); err != nil {
				log.Printf("❌ Pulse Monitor error: %v", err)
			}
		}
	}
}

// pulseMonitor fetches channel stats and sends them to Telegram
func pulseMonitor(ctx context.Context, platform *YouTubePlatform, ui *TelegramUI) error {
	log.Println("📈 Running Pulse Monitor...")

	stats, err := platform.GetChannelStats(ctx)
	if err != nil {
		return err
	}

	message := formatChannelStats(stats)

	return ui.SendMessage(ctx, message)
}

// runSmartResponder runs the comment responder service on an interval
func runSmartResponder(ctx context.Context, platform *YouTubePlatform, ai *LLMClient, ui *TelegramUI, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Cache to track processed comments
	processedComments := make(map[string]bool)

	for {
		select {
		case <-ctx.Done():
			log.Println("💬 Smart Responder stopping...")
			return
		case <-ticker.C:
			if err := smartResponder(ctx, platform, ai, ui, processedComments); err != nil {
				log.Printf("❌ Smart Responder error: %v", err)
			}
		}
	}
}

// smartResponder checks for new comments and handles them
func smartResponder(ctx context.Context, platform *YouTubePlatform, ai *LLMClient, ui *TelegramUI, processed map[string]bool) error {
	log.Println("💬 Checking for new comments...")

	comments, err := platform.GetRecentComments(ctx, 10)
	if err != nil {
		return err
	}

	for _, comment := range comments {
		// Skip if already processed
		if processed[comment.ID] {
			continue
		}

		// Generate AI reply
		aiReply, err := ai.GenerateReply(ctx, comment.Text)
		if err != nil {
			log.Printf("Failed to generate reply for comment %s: %v", comment.ID, err)
			continue
		}

		// Format notification message
		message := formatCommentNotification(comment, aiReply)

		// Send to Telegram (in a real implementation, you'd add buttons here)
		if err := ui.SendMessage(ctx, message); err != nil {
			log.Printf("Failed to send notification: %v", err)
			continue
		}

		// Mark as processed
		processed[comment.ID] = true

		log.Printf("✅ Processed comment from %s", comment.Author)
	}

	return nil
}

// Helper formatting functions
func formatVideo(index int, video *Video) string {
	return fmt.Sprintf("%d. \"%s\" - %s\n", index, video.Title, formatViews(video.ViewCount))
}

func formatViews(count int64) string {
	if count >= 1000000 {
		return fmt.Sprintf("%.1fM views", float64(count)/1000000.0)
	}
	if count >= 1000 {
		return fmt.Sprintf("%.1fK views", float64(count)/1000.0)
	}
	return fmt.Sprintf("%d views", count)
}

func formatChannelStats(stats *ChannelStats) string {
	msg := fmt.Sprintf(`📊 Channel Pulse Monitor

👥 Subscribers: %s

🎬 Recent Videos Performance:
`,
		formatNumber(stats.SubscriberCount))

	if len(stats.RecentVideos) == 0 {
		msg += "\nNo recent videos found\n"
	} else {
		for i, video := range stats.RecentVideos {
			// Calculate views per day
			viewsPerDay := int64(0)
			if video.AgeInDays > 0 {
				viewsPerDay = video.ViewCount / int64(video.AgeInDays)
			}

			msg += fmt.Sprintf(`
%d. %s
   👀 %s views (%s/day) | 👍 %s | 💬 %s
   📅 %d days ago`,
				i+1,
				truncateTitle(video.Title, 60),
				formatNumber(video.ViewCount),
				formatNumber(viewsPerDay),
				formatNumber(video.LikeCount),
				formatNumber(video.CommentCount),
				video.AgeInDays)
		}
	}

	msg += fmt.Sprintf("\n\nUpdated: %s", time.Now().Format("Jan 02, 15:04 MST"))
	return msg
}

func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}

func formatNumber(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.2fM", float64(n)/1000000.0)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}

func formatCommentNotification(comment *Comment, aiReply string) string {
	return fmt.Sprintf(`📬 New Comment

👤 User: %s
💬 Comment:
"%s"

🤖 AI Suggested Reply:
"%s"

[Note: In full version, buttons would appear here for Confirm/Edit/Ignore]`,
		comment.Author,
		comment.Text,
		aiReply,
	)
}
