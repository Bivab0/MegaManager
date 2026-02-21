# MegaManager AI - Quick Start Guide

## 🚀 Get Started in 5 Minutes

### 1. Get Your API Credentials

#### YouTube API Key
1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project
3. Enable "YouTube Data API v3"
4. Create credentials → API Key
5. Copy the key

#### Telegram Bot
1. Open Telegram and message [@BotFather](https://t.me/botfather)
2. Send `/newbot` and follow the prompts
3. Copy your bot token
4. Message [@userinfobot](https://t.me/userinfobot) to get your Chat ID

#### OpenAI API Key
1. Go to [OpenAI Platform](https://platform.openai.com/api-keys)
2. Create a new API key
3. Copy the key

#### YouTube Channel ID
1. Go to [YouTube Studio](https://studio.youtube.com)
2. Click your channel icon → Settings
3. Copy your Channel ID from the URL or Advanced settings

### 2. Configure the Bot

```bash
# Copy the example config
cp .env.example .env

# Edit with your credentials
nano .env
```

Your `.env` should look like:
```env
YOUTUBE_API_KEY=AIzaSy...your_key
YOUTUBE_CHANNEL_ID=UC...your_channel_id
TELEGRAM_BOT_TOKEN=123456:ABC...your_token
TELEGRAM_CHAT_ID=123456789
LLM_API_ENDPOINT=https://api.openai.com/v1/chat/completions
LLM_API_KEY=sk-...your_openai_key
LLM_MODEL=gpt-4o-mini
```

### 3. Run the Bot

```bash
# Option 1: Run directly
go run main.go

# Option 2: Build and run
go build -o megamanager
./megamanager
```

### 4. What to Expect

You'll see logs like:
```
🚀 MegaManager AI starting...
✅ All services started successfully
📊 Trend Scout: every 720 minutes
📈 Pulse Monitor: every 180 minutes
💬 Smart Responder: every 5 minutes
```

In Telegram, you'll receive:
- **Trending videos** - Every 12 hours
- **Channel stats** - Every 3 hours
- **Comment notifications** - When new comments appear

### 5. Stop the Bot

Press `Ctrl+C` to stop gracefully.

---

## 🐛 Troubleshooting

**"YOUTUBE_API_KEY is required"**
- Make sure your `.env` file is in the same directory
- Check that you copied all values correctly

**"YouTube API error (status 403)"**
- Enable YouTube Data API v3 in Google Cloud Console
- Check your API key is correct

**"Telegram API error"**
- Verify bot token from @BotFather
- Make sure you've sent `/start` to your bot
- Check Chat ID is correct

**No messages in Telegram**
- Check bot token and chat ID
- Make sure bot is running without errors
- Check your API quotas

---

## 📊 What Each Service Does

### Trend Scout (Every 12 hours)
Fetches the top 5 trending videos **in your niche** (same category as your channel) and sends them to you. Excludes your own videos. Get relevant content inspiration without sorting through unrelated viral content.

### Pulse Monitor (Every 3 hours)
Shows **actionable recent metrics**:
- Your last 5 videos' performance
- Views per day rate (spot momentum!)
- Likes and comments engagement
- Age of each video in days

No useless totals like "lifetime views" - only metrics that help you make decisions NOW.

### Smart Responder (Every 5 minutes)
Polls your videos for new comments, generates AI-suggested replies, and sends them to Telegram for your review.

---

## ⚙️ Customizing Intervals

Edit `.env` to change polling frequencies:

```env
TREND_SCOUT_INTERVAL=720      # Minutes (720 = 12 hours)
PULSE_MONITOR_INTERVAL=180    # Minutes (180 = 3 hours)
COMMENT_POLL_INTERVAL=5       # Minutes
```

---

## 📱 Example Telegram Messages

**Trending Videos in Your Niche:**
```
🔥 Trending in Your Niche (Top 5):

1. "How to build AI agents in 2026" - 2.3M views
2. "Golang microservices tutorial" - 1.8M views
...

💡 Inspiration: Pick a topic and create your version!
```

**Recent Video Performance:**
```
📊 Channel Pulse Monitor

👥 Subscribers: 12.5K

🎬 Recent Videos Performance:

1. Your Latest Video Title
   👀 15.2K views (3.8K/day) | 👍 892 | 💬 143
   📅 4 days ago

2. Previous Video Title
   👀 8.5K views (2.1K/day) | 👍 445 | 💬 67
   📅 4 days ago
...
```

**New Comment:**
```
📬 New Comment

👤 User: john_doe
💬 Comment: "Great video!"

🤖 AI Suggested Reply:
"Thanks for watching! Glad you enjoyed it."
```

---

That's it! Your bot is now running and monitoring your YouTube channel. 🎉
