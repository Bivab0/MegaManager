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

// Cache size limits to prevent memory growth over long runs
const (
	maxSeenVideos       = 500  // Max trending video IDs to track (~5KB)
	maxProcessedComments = 1000 // Max comment IDs to track (~25KB)
)

// Config holds all application configuration
type Config struct {
	// YouTube
	YouTubeAPIKey    string
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

	// Trend Scout settings
	TrendScoutCategory string // YouTube category ID (e.g., "20" for Gaming, "28" for Science & Tech)
	TrendScoutRegion   string // Region code (e.g., "US", "IN", "GB")

	// Feature toggles
	EnableTrendScout     bool
	EnablePulseMonitor   bool
	EnableSmartResponder bool
	EnableCommandHandler bool
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
		TrendScoutInterval:   getEnvInt("TREND_SCOUT_INTERVAL", 30), // Default 30 min for faster detection
		PulseMonitorInterval: getEnvInt("PULSE_MONITOR_INTERVAL", 180),
		CommentPollInterval:  getEnvInt("COMMENT_POLL_INTERVAL", 5),
		// Trend Scout settings
		TrendScoutCategory: os.Getenv("TREND_SCOUT_CATEGORY"), // Empty = auto-detect from channel
		TrendScoutRegion:   getEnv("TREND_SCOUT_REGION", "US"),
		// Feature toggles (all enabled by default)
		EnableTrendScout:     getEnvBool("ENABLE_TREND_SCOUT", true),
		EnablePulseMonitor:   getEnvBool("ENABLE_PULSE_MONITOR", true),
		EnableSmartResponder: getEnvBool("ENABLE_SMART_RESPONDER", true),
		EnableCommandHandler: getEnvBool("ENABLE_COMMAND_HANDLER", true),
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
	// LLM credentials only required if Smart Responder is enabled
	if cfg.EnableSmartResponder {
		if cfg.LLMEndpoint == "" {
			log.Fatal("LLM_API_ENDPOINT is required when ENABLE_SMART_RESPONDER is true")
		}
		if cfg.LLMAPIKey == "" {
			log.Fatal("LLM_API_KEY is required when ENABLE_SMART_RESPONDER is true")
		}
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

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	// Accept common boolean representations
	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func main() {
	log.Println("🚀 MegaManager AI starting...")

	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize providers
	platform := NewYouTubePlatform(cfg.YouTubeAPIKey, cfg.YouTubeChannelID, cfg.TrendScoutCategory, cfg.TrendScoutRegion)
	ui := NewTelegramUI(cfg.TelegramBotToken, cfg.TelegramChatID)

	// Only initialize LLM client if Smart Responder is enabled
	var ai *LLMClient
	if cfg.EnableSmartResponder {
		ai = NewLLMClient(cfg.LLMEndpoint, cfg.LLMAPIKey, cfg.LLMModel)
	}

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// WaitGroup to track all goroutines
	var wg sync.WaitGroup

	// Start Trend Scout (runs every N minutes)
	if cfg.EnableTrendScout {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runTrendScout(ctx, platform, ui, time.Duration(cfg.TrendScoutInterval)*time.Minute)
		}()
	}

	// Start Pulse Monitor (runs every N minutes)
	if cfg.EnablePulseMonitor {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runPulseMonitor(ctx, platform, ui, time.Duration(cfg.PulseMonitorInterval)*time.Minute)
		}()
	}

	// Start Smart Responder (polls every N minutes)
	if cfg.EnableSmartResponder {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runSmartResponder(ctx, platform, ai, ui, time.Duration(cfg.CommentPollInterval)*time.Minute)
		}()
	}

	// Start Command Handler (responds to Telegram commands)
	if cfg.EnableCommandHandler {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runCommandHandler(ctx, platform, ui)
		}()
	}

	// Log enabled services
	log.Println("✅ Services started:")
	if cfg.EnableTrendScout {
		log.Printf("  🔥 Trend Scout: every %d minutes", cfg.TrendScoutInterval)
	} else {
		log.Println("  🔥 Trend Scout: disabled")
	}
	if cfg.EnablePulseMonitor {
		log.Printf("  📈 Pulse Monitor: every %d minutes", cfg.PulseMonitorInterval)
	} else {
		log.Println("  📈 Pulse Monitor: disabled")
	}
	if cfg.EnableSmartResponder {
		log.Printf("  💬 Smart Responder: every %d minutes", cfg.CommentPollInterval)
	} else {
		log.Println("  💬 Smart Responder: disabled")
	}
	if cfg.EnableCommandHandler {
		log.Println("  🤖 Command Handler: listening for /stats and /trending")
	} else {
		log.Println("  🤖 Command Handler: disabled")
	}

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
// It tracks seen videos and only notifies when NEW videos enter trending
func runTrendScout(ctx context.Context, platform *YouTubePlatform, ui *TelegramUI, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Track seen video IDs to detect new trending videos
	seenVideos := make(map[string]bool)
	isFirstRun := true

	// Run immediately on start
	if err := trendScout(ctx, platform, ui, seenVideos, isFirstRun); err != nil {
		log.Printf("❌ Trend Scout error: %v", err)
	}
	isFirstRun = false

	for {
		select {
		case <-ctx.Done():
			log.Println("🔥 Trend Scout stopping...")
			return
		case <-ticker.C:
			// Prevent unbounded cache growth - clear if too large
			if len(seenVideos) > maxSeenVideos {
				log.Printf("🔥 Clearing trend cache (was %d entries)", len(seenVideos))
				seenVideos = make(map[string]bool)
				isFirstRun = true // Treat as first run to repopulate
			}

			if err := trendScout(ctx, platform, ui, seenVideos, isFirstRun); err != nil {
				log.Printf("❌ Trend Scout error: %v", err)
			}
			isFirstRun = false
		}
	}
}

// trendScout fetches trending videos and notifies only about NEW ones
func trendScout(ctx context.Context, platform *YouTubePlatform, ui *TelegramUI, seenVideos map[string]bool, isFirstRun bool) error {
	log.Println("🔥 Checking for new trending videos...")

	trending, err := platform.GetTrendingVideos(ctx, 10) // Fetch more to track
	if err != nil {
		return err
	}

	if len(trending) == 0 {
		log.Println("🔥 No trending videos found")
		return nil
	}

	// On first run, just populate the seen list and send a summary
	if isFirstRun {
		for _, video := range trending {
			seenVideos[video.ID] = true
		}

		// Send initial summary
		message := "`╔═══════════════════════════╗`\n"
		message += "`║ TREND SCOUT ACTIVE        ║`\n"
		message += "`╚═══════════════════════════╝`\n\n"
		message += fmt.Sprintf("📊 Tracking *%d* trending videos\n", len(trending))
		message += "🔔 You'll be notified when NEW videos trend\\!\n\n"
		message += "*Currently trending:*\n"
		for i, video := range trending {
			if i >= 5 {
				break
			}
			message += fmt.Sprintf("• %s \\(%s\\)\n", escapeMarkdown(truncateTitle(video.Title, 40)), formatViews(video.ViewCount))
		}

		return ui.SendMessage(ctx, message)
	}

	// Find NEW trending videos
	var newVideos []*Video
	for _, video := range trending {
		if !seenVideos[video.ID] {
			newVideos = append(newVideos, video)
			seenVideos[video.ID] = true
		}
	}

	if len(newVideos) == 0 {
		log.Println("🔥 No new trending videos detected")
		return nil
	}

	log.Printf("🔥 Found %d NEW trending video(s)!", len(newVideos))

	// Send notification for each new trending video with analytics
	for _, video := range newVideos {
		// Calculate analytics
		viewsPerHour := calculateViewsPerHour(video.ViewCount, video.PublishedAt)
		ageHours := calculateVideoAge(video.PublishedAt)
		heatEmoji, heatLevel := getHeatLevel(viewsPerHour)
		opportunity := getOpportunityWindow(viewsPerHour, ageHours)

		message := "`╔═══════════════════════════╗`\n"
		message += fmt.Sprintf("`║  %s TREND SURGING         ║`\n", heatEmoji)
		message += "`╚═══════════════════════════╝`\n\n"
		message += fmt.Sprintf("*%s*\n\n", escapeMarkdown(video.Title))
		message += fmt.Sprintf("👁 *%s* in %s\n", formatViewsCompact(video.ViewCount), formatAgeCompact(video.PublishedAt))
		message += fmt.Sprintf("📈 *%s/hour* \\(%s\\)\n", formatViewsCompact(viewsPerHour), heatLevel)
		message += fmt.Sprintf("⏱ Opportunity: %s\n\n", opportunity)
		message += fmt.Sprintf("🔗 https://youtube\\.com/watch?v=%s\n\n", video.ID)
		message += "💡 _Create your version while it's hot\\!_"

		if err := ui.SendMessage(ctx, message); err != nil {
			log.Printf("❌ Failed to send trending notification: %v", err)
		}
	}

	return nil
}

// getTrendingNow fetches and displays all current trending videos (for manual /trending command)
func getTrendingNow(ctx context.Context, platform *YouTubePlatform, ui *TelegramUI) error {
	log.Println("🔥 Fetching current trending videos...")

	trending, err := platform.GetTrendingVideos(ctx, 5)
	if err != nil {
		return err
	}

	message := "`╔═══════════════════════════╗`\n"
	message += "`║ TRENDING NOW              ║`\n"
	message += "`╚═══════════════════════════╝`\n\n"

	if len(trending) == 0 {
		message += "No trending videos found"
	} else {
		for i, video := range trending {
			viewsPerHour := calculateViewsPerHour(video.ViewCount, video.PublishedAt)
			heatEmoji, heatLevel := getHeatLevel(viewsPerHour)

			message += fmt.Sprintf("%s *#%d* \\[%s\\]\n", heatEmoji, i+1, heatLevel)
			message += "`───────────────────────────`\n"
			message += fmt.Sprintf("%s\n", escapeMarkdown(truncateTitle(video.Title, 45)))
			message += fmt.Sprintf("👁 %s in %s • 📈 %s/hr\n",
				formatViewsCompact(video.ViewCount),
				formatAgeCompact(video.PublishedAt),
				formatViewsCompact(viewsPerHour))
			message += fmt.Sprintf("🔗 https://youtube\\.com/watch?v=%s\n\n", video.ID)
		}
		message += "💡 *Pick a hot topic & create your version\\!*"
	}

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
			// Prevent unbounded cache growth - clear if too large
			if len(processedComments) > maxProcessedComments {
				log.Printf("💬 Clearing comment cache (was %d entries)", len(processedComments))
				processedComments = make(map[string]bool)
			}

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

// runCommandHandler listens for Telegram commands and responds
func runCommandHandler(ctx context.Context, platform *YouTubePlatform, ui *TelegramUI) {
	offset := 0
	ticker := time.NewTicker(2 * time.Second) // Poll every 2 seconds
	defer ticker.Stop()

	log.Println("🤖 Command Handler started")

	for {
		select {
		case <-ctx.Done():
			log.Println("🤖 Command Handler stopping...")
			return
		case <-ticker.C:
			updates, err := ui.GetUpdates(ctx, offset)
			if err != nil {
				// Only log non-timeout errors (timeouts are normal for long polling)
				if !strings.Contains(err.Error(), "context deadline exceeded") &&
					!strings.Contains(err.Error(), "timeout") {
					log.Printf("❌ Command Handler error: %v", err)
				}
				continue
			}

			for _, update := range updates {
				offset = update.UpdateID + 1

				// Handle messages with commands
				if update.Message != nil && update.Message.Text != "" {
					command := update.Message.Text

					switch command {
					case "/stats":
						log.Println("📊 Received /stats command")
						if err := pulseMonitor(ctx, platform, ui); err != nil {
							log.Printf("❌ Failed to send stats: %v", err)
						}

					case "/trending":
						log.Println("🔥 Received /trending command")
						if err := getTrendingNow(ctx, platform, ui); err != nil {
							log.Printf("❌ Failed to send trending: %v", err)
						}

					case "/start":
						welcomeMsg := "👋 *Welcome to MegaManager AI\\!*\n\n"
						welcomeMsg += "Available commands:\n"
						welcomeMsg += "• `/stats` \\- Get channel statistics\n"
						welcomeMsg += "• `/trending` \\- Get trending videos in your niche\n\n"
						welcomeMsg += "I'll also send you automatic updates:\n"
						welcomeMsg += "• 📊 Channel stats every 3 hours\n"
						welcomeMsg += "• 🔥 Trending videos every 12 hours\n"
						welcomeMsg += "• 💬 New comment notifications\\n"
						if err := ui.SendMessage(ctx, welcomeMsg); err != nil {
							log.Printf("❌ Failed to send welcome message: %v", err)
						}
					}
				}
			}
		}
	}
}

// Helper formatting functions
func escapeMarkdown(text string) string {
	// Escape Telegram Markdown special characters
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
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

// formatViewsCompact returns views without "views" suffix
func formatViewsCompact(count int64) string {
	if count >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(count)/1000000.0)
	}
	if count >= 1000 {
		return fmt.Sprintf("%.1fK", float64(count)/1000.0)
	}
	return fmt.Sprintf("%d", count)
}

// calculateVideoAge returns video age in hours
func calculateVideoAge(publishedAt time.Time) float64 {
	return time.Since(publishedAt).Hours()
}

// calculateViewsPerHour returns average views per hour
func calculateViewsPerHour(views int64, publishedAt time.Time) int64 {
	hours := calculateVideoAge(publishedAt)
	if hours < 1 {
		hours = 1 // Minimum 1 hour to avoid division issues
	}
	return int64(float64(views) / hours)
}

// formatAgeCompact returns a human-readable age string
func formatAgeCompact(publishedAt time.Time) string {
	hours := calculateVideoAge(publishedAt)
	if hours < 1 {
		return "< 1 hour"
	}
	if hours < 24 {
		return fmt.Sprintf("%.0f hours", hours)
	}
	days := hours / 24
	if days < 7 {
		return fmt.Sprintf("%.0f days", days)
	}
	return fmt.Sprintf("%.0f weeks", days/7)
}

// getHeatLevel returns a heat indicator based on views per hour
func getHeatLevel(viewsPerHour int64) (string, string) {
	// Heat levels based on views/hour velocity
	if viewsPerHour >= 50000 {
		return "🔥🔥🔥", "EXTREME"
	}
	if viewsPerHour >= 20000 {
		return "🔥🔥", "VERY HIGH"
	}
	if viewsPerHour >= 10000 {
		return "🔥", "HIGH"
	}
	if viewsPerHour >= 5000 {
		return "📈", "MODERATE"
	}
	return "📊", "NORMAL"
}

// getOpportunityWindow returns opportunity assessment
func getOpportunityWindow(viewsPerHour int64, ageHours float64) string {
	// Best opportunity: high velocity + young video
	if viewsPerHour >= 10000 && ageHours < 12 {
		return "🟢 HIGH \\- Act now\\!"
	}
	if viewsPerHour >= 5000 && ageHours < 24 {
		return "🟡 MEDIUM \\- Good timing"
	}
	if ageHours > 48 {
		return "🔴 LOW \\- Topic maturing"
	}
	return "🟡 MEDIUM"
}

func formatChannelStats(stats *ChannelStats) string {
	msg := "`╔═══════════════════════════╗`\n"
	msg += "`║  CHANNEL PULSE MONITOR    ║`\n"
	msg += "`╚═══════════════════════════╝`\n\n"
	msg += fmt.Sprintf("👥 Subscribers: *%s*\n\n", formatNumber(stats.SubscriberCount))

	if len(stats.RecentVideos) == 0 {
		msg += "No recent videos found"
		return msg
	}

	for i, video := range stats.RecentVideos {
		// Calculate views per day

		msg += fmt.Sprintf("\n\n*▶ VIDEO %d*\n", i+1)
		msg += "`───────────────────────────`\n"
		msg += fmt.Sprintf("%s\n", escapeMarkdown(truncateTitle(video.Title, 50)))
		msg += fmt.Sprintf("📅 %d days ago\n", video.AgeInDays)
		msg += fmt.Sprintf("👁 %s  👍 %s  💬 %s\n",
			formatNumberCompact(video.ViewCount),
			formatNumberCompact(video.LikeCount),
			formatNumberCompact(video.CommentCount))
	}

	msg += fmt.Sprintf("_Updated: %s_", time.Now().Format("Jan 02, 15:04"))

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

func formatNumberCompact(n int64) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000.0)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000.0)
	}
	return fmt.Sprintf("%d", n)
}

func formatCommentNotification(comment *Comment, aiReply string) string {
	msg := "`╔═══════════════════════════╗`\n"
	msg += "`║      NEW COMMENT          ║`\n"
	msg += "`╚═══════════════════════════╝`\n\n"

	msg += fmt.Sprintf("👤 *From:* %s\n\n", escapeMarkdown(comment.Author))

	msg += "*💬 COMMENT:*\n"
	msg += fmt.Sprintf("%s\n\n", escapeMarkdown(wrapText(comment.Text, 50)))

	msg += "*🤖 SUGGESTED REPLY:*\n"
	msg += fmt.Sprintf("%s\n\n", escapeMarkdown(wrapText(aiReply, 50)))

	msg += "⚠️ _Reply manually on YouTube_"

	return msg
}

func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return strings.Join(lines, "\n")
}
