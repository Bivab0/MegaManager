# MegaManager AI

> **Minimal YouTube automation bot in Go - 4 files, 970 lines, < 10MB RAM**

**MegaManager** is a lightweight, stateless automation engine built in pure Go (standard library only). It monitors your YouTube channel and sends intelligent notifications to Telegram, giving you "human-in-the-loop" control over your content strategy.

**Quick Start:**
```bash
cp .env.example .env     # Copy config file
nano .env                # Add your API keys
go run main.go           # Start the bot
```

## 🌟 Core Value Proposition

Most automation bots are either fully autonomous (risky) or fully manual (time-consuming). **MegaManager** finds the middle ground: it monitors your channel  and prepares everything for you, so you only need to tap a button on your phone to stay productive.

---

## 🛠️ Key Functional Features

### 1. The Trend Scout (Content Strategy)

* **Frequency:** Twice daily (e.g., Morning and Evening).
* **Action:** The bot automatically scans the platform’s trending charts.
* **Output:** It sends a curated list of **5 trending topics** directly to your Telegram.
* **Purpose:** To give you instant inspiration for your next video without you having to manually research what's "viral."

### 2. The Pulse Monitor (Channel Analytics)

* **Frequency:** Every 3 hours (180 minutes).
* **Action:** The bot performs a "Stat Snapshot." It calls the platform API to fetch current views, subscriber counts, and engagement metrics.
* **Output:** A clean, concise report sent to Telegram showing how your channel is performing at that exact moment.
* **Statelessness:** No database is used. It’s a "Live Feed" approach—it simply grabs the data and hands it to you.

### 3. The Smart Responder (Community Management)

* **Action:** The bot constantly polls for new comments on any of your videos.
* **AI Integration:** When a new comment is found, the bot sends the text to an LLM via an HTTP call.
* **The Approval Flow:**
* You receive a Telegram notification: *"User X commented: [Comment Text]"*.
* Below it, the bot provides an **AI-Suggested Reply**.
* **Decision Buttons:**
* **[✅ Confirm]:** The bot immediately posts the AI reply to the video.
* **[📝 Edit]:** You can type a custom message if the AI didn't get it quite right.
* **[❌ Ignore]:** Dismisses the notification.





---

## 🏗️ Architectural Philosophy

### 🔌 Provider-Agnostic (Modular)

The system is built using Go **Interfaces**. While you might start with YouTube, the code is structured so that you can swap "YouTube" for "Twitter," "Reddit," or "Discord" by simply changing the data provider module.

### 🧠 External Intelligence (HTTP-Based)

The bot does not run the AI locally. It communicates with an LLM (like OpenAI or a private server) via standard **HTTP POST** requests. This makes the bot extremely lightweight and easy to host on a small VPS.

### 💾 Stateless & DB-Less

To keep the system fast and "zero-maintenance," there is **no database**.

* It uses **In-Memory Caching** for short-term comparisons.
* It uses **Telegram Callback Data** to "remember" which comment you are replying to when you click a button.

---

## 🚀 Quick Start

### Prerequisites

* **Go 1.21+** installed on your system
* **YouTube Data API v3** credentials (API Key)
* **Telegram Bot Token** (from [@BotFather](https://t.me/botfather))
* **LLM API Access** (OpenAI API key or compatible endpoint)

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/megamanager.git
cd megamanager

# Install dependencies
go mod download

# Copy environment configuration
cp .env.example .env

# Edit .env with your API credentials
nano .env
```

### Configuration

Create a `.env` file with the following variables:

```env
# YouTube API
YOUTUBE_API_KEY=your_youtube_api_key_here
YOUTUBE_CHANNEL_ID=your_channel_id_here

# Telegram Bot
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here
TELEGRAM_CHAT_ID=your_telegram_chat_id_here

# LLM Provider
LLM_API_ENDPOINT=https://api.openai.com/v1/chat/completions
LLM_API_KEY=your_llm_api_key_here
LLM_MODEL=gpt-4

# Polling Intervals (in minutes)
TREND_SCOUT_INTERVAL=720    # 12 hours
PULSE_MONITOR_INTERVAL=180  # 3 hours
COMMENT_POLL_INTERVAL=5     # 5 minutes
```

### Running the Bot

```bash
# Run directly
go run main.go

# Or build and run
go build -o megamanager
./megamanager
```

---

## 📁 Project Structure

**Minimal & Lightweight Design - Just 4 Files!**

```
megamanager/
├── main.go         # The Orchestrator (timers, loops, coordination)
├── platform.go     # YouTube data provider (API client)
├── ai.go           # LLM HTTP client (OpenAI-compatible)
├── ui.go           # Telegram UI & interaction logic
├── go.mod          # Dependencies
├── .env.example    # Example environment file
├── .gitignore      # Git ignore file
├── README.md       # This file
└── PLAN.md         # Implementation roadmap
```

**Design Philosophy:**
- **Minimal Dependencies:** Uses only Go standard library (`net/http`, `encoding/json`)
- **Small Footprint:** Runs in under 10MB of RAM
- **Simple Architecture:** No frameworks, no complex abstractions
- **Easy to Understand:** 4 files, ~1000 lines of code total

---

## 🔧 Technology Stack

* **Language:** Go 1.21+ (pure standard library)
* **HTTP Client:** `net/http` (zero external dependencies)
* **JSON:** `encoding/json` (standard library)
* **Telegram Bot API:** Direct REST API calls
* **YouTube API:** Direct REST API calls (no SDK needed)
* **LLM Integration:** HTTP client for OpenAI-compatible endpoints
* **Scheduler:** Go `time.Ticker` for interval-based execution
* **Caching:** Simple in-memory `map[string]bool` for deduplication
* **Memory Footprint:** < 10MB RAM usage

---

## 🎯 Usage Examples

### Trend Scout Notification

When the bot runs its trend scan, you'll receive a Telegram message like:

```
🔥 Trending Now (Top 5):

1. "How to build AI agents" - 2.3M views
2. "Golang vs Rust comparison" - 1.8M views
3. "Best productivity apps 2026" - 1.5M views
4. "Cloud architecture patterns" - 1.2M views
5. "API design best practices" - 980K views

💡 Pick one and create your next video!
```

### Smart Responder Flow

```
📬 New Comment on "Your Video Title"

👤 User: @john_doe
💬 "Great tutorial! Can you make one about microservices?"

🤖 AI Suggested Reply:
"Thanks for watching! That's a great suggestion. I'll add microservices to my content roadmap. Stay tuned!"

[✅ Confirm] [📝 Edit] [❌ Ignore]
```

---

## 🛣️ Roadmap

See [PLAN.md](PLAN.md) for the detailed implementation plan.

**Phase 1:** Setup & Basic Structure (1-2 days)
* ✅ Project structure created
* ✅ Configuration system (environment variables)
* ✅ Basic file skeleton

**Phase 2:** YouTube Integration (2-3 days)
* Implement YouTube API client (`platform.go`)
* Test trending videos, channel stats, comments

**Phase 3:** LLM Integration (1-2 days)
* Implement OpenAI client (`ai.go`)
* Test comment reply generation

**Phase 4:** Telegram Integration (2-3 days)
* Implement Telegram bot (`ui.go`)
* Test message sending and callbacks

**Phase 5:** Orchestration & Testing (2-3 days)
* Wire up all services in `main.go`
* End-to-end testing
* Deploy and monitor

---

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

---

## 🐛 Troubleshooting

### "Invalid API Key" Error
* Verify your API keys in `.env` are correct
* Check that the YouTube API is enabled in Google Cloud Console
* Ensure billing is set up for YouTube API quota

### Bot Not Responding to Telegram
* Verify your bot token with [@BotFather](https://t.me/botfather)
* Check that `TELEGRAM_CHAT_ID` is your actual chat ID (use [@userinfobot](https://t.me/userinfobot))
* Ensure the bot has been started with `/start` command

### Comments Not Being Detected
* YouTube API has quota limits - check your usage in Google Cloud Console
* Ensure your channel has comment notifications enabled
* Verify `YOUTUBE_CHANNEL_ID` is correct

---

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

## 📧 Support

* **Issues:** [GitHub Issues](https://github.com/yourusername/megamanager/issues)
* **Email:** your.email@example.com
* **Telegram:** @yourusername

---

## 🙏 Acknowledgments

* YouTube Data API by Google
* Telegram Bot API
* OpenAI for LLM capabilities
* The Go community for excellent tooling

---

**Built with ❤️ for content creators who want to stay engaged without being glued to their screens.**

