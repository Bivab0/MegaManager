# MegaManager AI - Minimal Implementation Plan

## 🎯 Project Overview

**Goal:** Build a minimal, stateless automation engine in Go that provides human-in-the-loop automation for YouTube content creators.

**Architecture:** 4 files, zero external dependencies, < 10MB RAM

**Timeline:** 10-14 days

**Key Principles:**
- Simplicity over complexity
- Standard library only
- Stateless (no database, just in-memory caching)
- Direct API calls (no SDKs)
- Memory efficient

---

## 📦 File Structure

```
megamanager/
├── main.go         # Orchestrator: timers, scheduling, service coordination
├── platform.go     # YouTube API client: trends, stats, comments
├── ai.go           # LLM client: OpenAI-compatible HTTP calls
├── ui.go           # Telegram bot: messages, buttons, callbacks
├── go.mod          # Module definition (no external deps)
├── .env.example    # Configuration template
└── .gitignore      # Git ignore rules
```

---

## 🚀 Implementation Phases

### Phase 1: Project Setup (Day 1) ✅

**Status: COMPLETED**

- [x] Initialize Go module
- [x] Create directory structure
- [x] Create `.env.example` with all required variables
- [x] Create `.gitignore`
- [x] Set up basic file skeletons

**Files Created:**
- `main.go` - Configuration loading, service orchestration
- `platform.go` - YouTube API client skeleton
- `ai.go` - LLM client skeleton
- `ui.go` - Telegram client skeleton

---

### Phase 2: YouTube Integration (Days 2-4)

**Goal:** Implement all YouTube API calls in `platform.go`

**Tasks:**

#### 2.1 Trending Videos API
- [ ] Implement `GetTrendingVideos(ctx, limit)` function
- [ ] Parse YouTube API response
- [ ] Extract video ID, title, view count
- [ ] Test with real API key
- [ ] Handle API errors gracefully

**Testing:**
```bash
# Create test file
echo "YOUTUBE_API_KEY=your_key" > .env
go run main.go
```

#### 2.2 Channel Statistics API
- [ ] Implement `GetChannelStats(ctx)` function
- [ ] Fetch subscriber count, total views, video count
- [ ] Parse and return structured data
- [ ] Test with real channel ID

#### 2.3 Comments API
- [ ] Implement `GetRecentComments(ctx, limit)` function
- [ ] First: Get recent videos from channel
- [ ] Then: Get comments from those videos
- [ ] Parse comment data (ID, author, text, timestamp)
- [ ] Handle pagination if needed
- [ ] Test comment retrieval

**Deliverable:** Fully working YouTube client that can fetch all required data.

---

### Phase 3: LLM Integration (Days 5-6)

**Goal:** Implement OpenAI-compatible LLM client in `ai.go`

**Tasks:**

#### 3.1 HTTP Client Setup
- [ ] Implement `GenerateReply(ctx, commentText)` function
- [ ] Create request structure (messages, model, temperature)
- [ ] Add proper headers (Authorization, Content-Type)
- [ ] Set timeout (30 seconds)

#### 3.2 Prompt Engineering
- [ ] Design system prompt for YouTube replies
- [ ] Keep replies concise (1-2 sentences)
- [ ] Make replies friendly and engaging
- [ ] Test with various comment types

#### 3.3 Error Handling
- [ ] Handle API errors (rate limits, invalid keys)
- [ ] Implement retry logic with exponential backoff
- [ ] Log failures clearly
- [ ] Return fallback message on failure

**Testing:**
```bash
# Test with real comment
export LLM_API_KEY=your_openai_key
go run main.go
```

**Deliverable:** Working LLM client that generates appropriate replies.

---

### Phase 4: Telegram Integration (Days 7-9)

**Goal:** Implement Telegram bot in `ui.go`

**Tasks:**

#### 4.1 Basic Message Sending
- [ ] Implement `SendMessage(ctx, text)` function
- [ ] Support Markdown formatting
- [ ] Handle Telegram API errors
- [ ] Test message delivery

#### 4.2 Interactive Buttons (Optional for MVP)
- [ ] Implement `SendMessageWithButtons(ctx, text, buttons)` function
- [ ] Create inline keyboard structure
- [ ] Test button rendering

#### 4.3 Callback Handling (Optional for MVP)
- [ ] Implement `GetUpdates(ctx, offset)` function
- [ ] Poll for button callbacks
- [ ] Parse callback data
- [ ] Handle user actions (Confirm/Edit/Ignore)

**Note:** For MVP, we can skip buttons and just send notifications. User manually replies to comments on YouTube.

**Testing:**
```bash
# Get your chat ID
# 1. Message @userinfobot on Telegram
# 2. Copy your chat ID
export TELEGRAM_BOT_TOKEN=your_bot_token
export TELEGRAM_CHAT_ID=your_chat_id
go run main.go
```

**Deliverable:** Working Telegram integration that sends formatted messages.

---

### Phase 5: Service Implementation (Days 10-11)

**Goal:** Implement the three core services in `main.go`

#### 5.1 Trend Scout Service
- [ ] Implement `trendScout()` function
- [ ] Fetch top 5 trending videos
- [ ] Format as Telegram message
- [ ] Send to user
- [ ] Run on timer (every 12 hours)

**Output Format:**
```
🔥 Trending Now (Top 5):

1. "Video Title" - 2.3M views
2. "Video Title" - 1.8M views
3. "Video Title" - 1.5M views
4. "Video Title" - 1.2M views
5. "Video Title" - 980K views

💡 Pick one and create your next video!
```

#### 5.2 Pulse Monitor Service
- [ ] Implement `pulseMonitor()` function
- [ ] Fetch channel statistics
- [ ] Format analytics report
- [ ] Send to user
- [ ] Run on timer (every 3 hours)

**Output Format:**
```
📊 Channel Analytics

👥 Subscribers: 12.5K
👀 Total Views: 1.2M
🎬 Total Videos: 45

Updated: Jan 21, 2026 15:04 PST
```

#### 5.3 Smart Responder Service
- [ ] Implement `smartResponder()` function
- [ ] Poll for new comments (every 5 minutes)
- [ ] Use in-memory cache to track processed comments
- [ ] For each new comment:
  - [ ] Generate AI reply
  - [ ] Send notification to Telegram
  - [ ] Mark comment as processed
- [ ] Handle errors gracefully

**Output Format (MVP - no buttons):**
```
📬 New Comment

👤 User: @john_doe
💬 Comment:
"Great tutorial! Can you make one about microservices?"

🤖 AI Suggested Reply:
"Thanks for watching! That's a great suggestion. I'll add microservices to my content roadmap. Stay tuned!"

[Note: Go to YouTube to reply manually]
```

#### 5.4 Service Orchestration
- [ ] Set up context for graceful shutdown
- [ ] Run all services in goroutines
- [ ] Use `time.Ticker` for scheduling
- [ ] Handle OS signals (Ctrl+C)
- [ ] Implement WaitGroup for clean shutdown

**Deliverable:** All three services running concurrently on schedule.

---

### Phase 6: Testing & Debugging (Day 12)

**Goal:** End-to-end testing and bug fixes

**Testing Checklist:**

- [ ] Test with invalid API keys (should fail gracefully)
- [ ] Test with rate limits (should handle errors)
- [ ] Test with no comments (should not crash)
- [ ] Test with malformed API responses
- [ ] Test graceful shutdown (Ctrl+C)
- [ ] Monitor memory usage (should be < 10MB)
- [ ] Check for goroutine leaks
- [ ] Test all three services working together

**Tools:**
```bash
# Monitor memory
go build -o megamanager
./megamanager &
ps aux | grep megamanager

# Check goroutines
kill -QUIT <pid>  # Prints goroutine stack trace

# Test graceful shutdown
kill -TERM <pid>
```

---

### Phase 7: Polish & Documentation (Day 13)

**Tasks:**

- [ ] Add detailed logging for debugging
- [ ] Improve error messages
- [ ] Add code comments
- [ ] Update README with:
  - [ ] Setup instructions
  - [ ] API key acquisition guides
  - [ ] Troubleshooting section
- [ ] Create example `.env.example`
- [ ] Write deployment guide

---

### Phase 8: Deployment (Day 14)

**Goal:** Deploy to production

#### 8.1 VPS Setup (Option 1: DigitalOcean)
```bash
# SSH into VPS
ssh root@your-vps-ip

# Install Go
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone repo
git clone https://github.com/yourusername/megamanager.git
cd megamanager

# Set up environment
nano .env
# Paste your API keys

# Build
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
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl enable megamanager
sudo systemctl start megamanager

# Check status
sudo systemctl status megamanager

# View logs
sudo journalctl -u megamanager -f
```

#### 8.2 Docker Deployment (Option 2)
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
# Build and run
docker build -t megamanager .
docker run -d --name megamanager --restart=always megamanager
```

---

## ✅ Implementation Checklist

### Setup Phase
- [x] Initialize Go module
- [x] Create file structure
- [x] Set up configuration loading
- [x] Create .env.example

### YouTube Integration
- [x] Skeleton implementation
- [ ] Test trending videos API
- [ ] Test channel stats API
- [ ] Test comments API
- [ ] Handle all error cases

### LLM Integration
- [x] Skeleton implementation
- [ ] Test with OpenAI
- [ ] Test with other providers
- [ ] Optimize prompts
- [ ] Add retry logic

### Telegram Integration
- [x] Skeleton implementation
- [ ] Test message sending
- [ ] Test markdown formatting
- [ ] (Optional) Add button support
- [ ] (Optional) Add callback handling

### Services
- [x] Skeleton implementation
- [ ] Implement Trend Scout
- [ ] Implement Pulse Monitor
- [ ] Implement Smart Responder
- [ ] Test all services together

### Polish
- [ ] Add comprehensive logging
- [ ] Improve error handling
- [ ] Add code documentation
- [ ] Update README
- [ ] Create deployment guide

### Deployment
- [ ] Choose hosting provider
- [ ] Set up VPS
- [ ] Deploy application
- [ ] Set up monitoring
- [ ] Test in production

---

## 📊 Progress Tracking

| Phase | Status | Estimated | Actual |
|-------|--------|-----------|--------|
| 1. Setup | ✅ Done | 1 day | 1 day |
| 2. YouTube | 🔄 In Progress | 3 days | - |
| 3. LLM | ⏳ Pending | 2 days | - |
| 4. Telegram | ⏳ Pending | 3 days | - |
| 5. Services | ⏳ Pending | 2 days | - |
| 6. Testing | ⏳ Pending | 1 day | - |
| 7. Polish | ⏳ Pending | 1 day | - |
| 8. Deploy | ⏳ Pending | 1 day | - |

---

## 🎯 MVP Features (Must Have)

1. ✅ Configuration from environment variables
2. ✅ YouTube trending videos (Trend Scout)
3. ✅ YouTube channel stats (Pulse Monitor)
4. ✅ YouTube comment detection
5. ✅ AI reply generation
6. ✅ Telegram notifications
7. [ ] All services working together
8. [ ] Graceful shutdown
9. [ ] Error handling

---

## 🚀 Future Enhancements (Nice to Have)

- [ ] Interactive buttons for comment approval
- [ ] Multi-platform support (Twitter, Reddit)
- [ ] Web dashboard
- [ ] Sentiment analysis
- [ ] Auto-reply for simple comments
- [ ] Comment analytics
- [ ] Video performance predictions

---

## 📝 Development Notes

### Current Status
- Project structure created
- All file skeletons in place
- Configuration system working
- Ready to implement YouTube API integration

### Next Steps
1. Get YouTube API key from Google Cloud Console
2. Test trending videos API
3. Test channel stats API
4. Test comments API
5. Move to LLM integration

### Blockers
None currently

### Decisions Made
1. **No external dependencies:** Keep it simple with stdlib only
2. **No database:** Use in-memory caching for deduplication
3. **Direct API calls:** No SDKs to keep binary small
4. **Polling over webhooks:** Simpler to set up and maintain
5. **MVP first:** Get basic functionality working before adding buttons

---

**Last Updated:** 2026-02-21
**Status:** Phase 1 Complete, Phase 2 Ready to Start
**Next Action:** Implement and test YouTube API integration
