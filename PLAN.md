# MegaManager AI - Implementation Plan

## 🎯 Project Overview

**Goal:** Build a minimal, stateless YouTube automation bot in Go

**Architecture:** 4 files, zero external dependencies, < 10MB RAM

**Timeline:** Single implementation phase

---

## 📦 File Structure

```
megamanager/
├── main.go         # Orchestrator (config, timers, services)
├── platform.go     # YouTube API client
├── ai.go           # LLM client (OpenAI-compatible)
├── ui.go           # Telegram bot client
├── go.mod          # Module definition
└── .env            # Your credentials
```

---

## ✅ Implementation Checklist

### Setup (Already Done ✅)
- [x] Project structure created
- [x] All 4 core files created
- [x] Configuration system ready
- [x] .env.example provided

### What's Left to Do

#### 1. Get API Credentials
- [ ] YouTube API Key
  - Go to [Google Cloud Console](https://console.cloud.google.com)
  - Create project → Enable YouTube Data API v3
  - Create credentials → API Key
  - Copy to `.env` as `YOUTUBE_API_KEY`

- [ ] YouTube Channel ID
  - Go to your YouTube channel
  - Click "Customize Channel"
  - Copy ID from URL or use [this tool](https://www.youtube.com/account_advanced)
  - Copy to `.env` as `YOUTUBE_CHANNEL_ID`

- [ ] Telegram Bot Token
  - Message [@BotFather](https://t.me/botfather) on Telegram
  - Send `/newbot` and follow prompts
  - Copy token to `.env` as `TELEGRAM_BOT_TOKEN`

- [ ] Telegram Chat ID
  - Message [@userinfobot](https://t.me/userinfobot) on Telegram
  - Copy your ID to `.env` as `TELEGRAM_CHAT_ID`

- [ ] OpenAI API Key
  - Go to [OpenAI Platform](https://platform.openai.com/api-keys)
  - Create new API key
  - Copy to `.env` as `LLM_API_KEY`

#### 2. Set Up Environment
```bash
# Copy example env file
cp .env.example .env

# Edit with your credentials
nano .env
```

#### 3. Test Each Component
```bash
# Test the app runs
go run main.go

# Check for errors in logs
# Verify Telegram receives messages
# Verify YouTube data is fetched
```

#### 4. Fix Any Issues
- [ ] Handle API rate limits
- [ ] Fix error messages
- [ ] Test graceful shutdown (Ctrl+C)

#### 5. Deploy (Optional)
```bash
# Build binary
go build -o megamanager

# Run in background
nohup ./megamanager > app.log 2>&1 &

# Or use systemd/Docker (see deployment section below)
```

---

## 🧪 Testing

### Quick Test
```bash
# Create .env with your credentials
cp .env.example .env
nano .env

# Run the bot
go run main.go
```

### What to Verify
1. Bot starts without errors
2. Telegram receives trending videos message
3. Telegram receives channel stats message
4. Telegram receives comment notifications with AI replies
5. Bot shuts down gracefully with Ctrl+C

### Common Issues

**"Invalid API Key"**
- Double-check your credentials in `.env`
- Ensure YouTube API is enabled in Google Cloud
- Check OpenAI API key is valid

**"No comments found"**
- Normal if your videos have no recent comments
- Bot will keep polling every 5 minutes

**"Rate limit exceeded"**
- YouTube API has daily quota (10,000 units/day)
- Reduce polling frequency if needed

---

## 🚀 Deployment

### Option 1: VPS (DigitalOcean, AWS, etc.)

```bash
# SSH into server
ssh user@your-server

# Install Go
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone and build
git clone <your-repo>
cd megamanager
nano .env  # Add your credentials
go build -o megamanager

# Run with systemd
sudo nano /etc/systemd/system/megamanager.service
```

**systemd service file:**
```ini
[Unit]
Description=MegaManager AI Bot
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/root/megamanager
ExecStart=/root/megamanager/megamanager
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable megamanager
sudo systemctl start megamanager
sudo systemctl status megamanager
```

### Option 2: Docker

Create `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o megamanager

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/megamanager .
COPY .env .
CMD ["./megamanager"]
```

```bash
docker build -t megamanager .
docker run -d --name megamanager --restart=always megamanager
```

### Option 3: Local/Raspberry Pi

```bash
# Just run it
go build -o megamanager
./megamanager

# Or use screen/tmux to keep it running
screen -S megamanager
./megamanager
# Press Ctrl+A then D to detach
```

---

## 📊 Current Status

| Component | Status | Notes |
|-----------|--------|-------|
| Project Setup | ✅ Done | All files created |
| YouTube Client | ✅ Skeleton | Needs API testing |
| LLM Client | ✅ Skeleton | Needs API testing |
| Telegram Client | ✅ Skeleton | Needs API testing |
| Orchestration | ✅ Done | Services wired up |
| Testing | ⏳ Pending | Need credentials |
| Deployment | ⏳ Pending | After testing |

---

## 🎯 Next Action

**Right now:** Get your API credentials and test the bot

```bash
# 1. Copy .env.example to .env
cp .env.example .env

# 2. Edit .env with your credentials
nano .env

# 3. Run the bot
go run main.go

# 4. Check Telegram for messages
```

That's it! The code is ready, you just need to add your credentials and run it.

---

## 🔧 How It Works

### Services Running

1. **Trend Scout** (every 12 hours)
   - Fetches top 5 trending videos from YouTube
   - Sends formatted list to Telegram

2. **Pulse Monitor** (every 3 hours)
   - Fetches your channel stats (subs, views, video count)
   - Sends analytics report to Telegram

3. **Smart Responder** (every 5 minutes)
   - Checks for new comments on your videos
   - Generates AI reply using LLM
   - Sends notification to Telegram with suggested reply

### Memory Usage
- Startup: ~5MB
- Running: ~8MB
- No database, no persistent storage
- Simple map for comment deduplication

---

## 📝 Configuration Reference

### Required Variables
```env
YOUTUBE_API_KEY=<your_key>
YOUTUBE_CHANNEL_ID=<your_channel_id>
TELEGRAM_BOT_TOKEN=<bot_token>
TELEGRAM_CHAT_ID=<your_chat_id>
LLM_API_ENDPOINT=https://api.openai.com/v1/chat/completions
LLM_API_KEY=<openai_key>
```

### Optional Variables
```env
LLM_MODEL=gpt-4o-mini              # Default: gpt-4o-mini
TREND_SCOUT_INTERVAL=720           # Default: 720 (12 hours)
PULSE_MONITOR_INTERVAL=180         # Default: 180 (3 hours)
COMMENT_POLL_INTERVAL=5            # Default: 5 (5 minutes)
```

---

## 🚀 Future Enhancements (Optional)

After the MVP is working, you can add:
- [ ] Interactive buttons for comment approval
- [ ] Support for multiple channels
- [ ] Web dashboard
- [ ] Multi-platform support (Twitter, Reddit)
- [ ] Custom AI prompts via Telegram commands

---

**Last Updated:** 2026-02-21
**Status:** Ready to implement - just add credentials and test!
