package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TelegramUI handles all Telegram bot interactions
type TelegramUI struct {
	botToken string
	chatID   string
	client   *http.Client
	baseURL  string
}

// NewTelegramUI creates a new Telegram UI handler
func NewTelegramUI(botToken, chatID string) *TelegramUI {
	return &TelegramUI{
		botToken: botToken,
		chatID:   chatID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", botToken),
	}
}

// SendMessage sends a text message to Telegram
func (t *TelegramUI) SendMessage(ctx context.Context, text string) error {
	endpoint := fmt.Sprintf("%s/sendMessage", t.baseURL)

	requestBody := map[string]interface{}{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Telegram API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SendMessageWithButtons sends a message with inline keyboard buttons
func (t *TelegramUI) SendMessageWithButtons(ctx context.Context, text string, buttons [][]InlineButton) (int, error) {
	endpoint := fmt.Sprintf("%s/sendMessage", t.baseURL)

	keyboard := make([][]map[string]interface{}, len(buttons))
	for i, row := range buttons {
		keyboard[i] = make([]map[string]interface{}, len(row))
		for j, btn := range row {
			keyboard[i][j] = map[string]interface{}{
				"text":          btn.Text,
				"callback_data": btn.CallbackData,
			}
		}
	}

	requestBody := map[string]interface{}{
		"chat_id": t.chatID,
		"text":    text,
		"reply_markup": map[string]interface{}{
			"inline_keyboard": keyboard,
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("Telegram API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Result.MessageID, nil
}

// GetUpdates retrieves new updates from Telegram (for callback handling)
func (t *TelegramUI) GetUpdates(ctx context.Context, offset int) ([]Update, error) {
	endpoint := fmt.Sprintf("%s/getUpdates", t.baseURL)

	requestBody := map[string]interface{}{
		"offset":  offset,
		"timeout": 25, // Telegram long polling timeout (keep under client timeout)
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Telegram API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Result []Update `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Result, nil
}

// InlineButton represents a Telegram inline keyboard button
type InlineButton struct {
	Text         string
	CallbackData string
}

// Update represents a Telegram update
type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message,omitempty"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

// Message represents a Telegram message
type Message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
	From      *User  `json:"from"`
}

// CallbackQuery represents a callback from an inline button
type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

// User represents a Telegram user
type User struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
}
