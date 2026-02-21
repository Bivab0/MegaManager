package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Video represents a YouTube video
type Video struct {
	ID         string
	Title      string
	ViewCount  int64
	ChannelID  string
	PublishedAt time.Time
}

// Comment represents a YouTube comment
type Comment struct {
	ID        string
	VideoID   string
	Author    string
	Text      string
	CreatedAt time.Time
}

// ChannelStats represents channel analytics
type ChannelStats struct {
	SubscriberCount int64
	ViewCount       int64
	VideoCount      int64
}

// YouTubePlatform handles all YouTube API operations
type YouTubePlatform struct {
	apiKey    string
	channelID string
	client    *http.Client
	baseURL   string
}

// NewYouTubePlatform creates a new YouTube platform provider
func NewYouTubePlatform(apiKey, channelID string) *YouTubePlatform {
	return &YouTubePlatform{
		apiKey:    apiKey,
		channelID: channelID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://www.googleapis.com/youtube/v3",
	}
}

// GetTrendingVideos fetches trending videos
func (y *YouTubePlatform) GetTrendingVideos(ctx context.Context, limit int) ([]*Video, error) {
	endpoint := fmt.Sprintf("%s/videos", y.baseURL)

	params := url.Values{}
	params.Set("part", "snippet,statistics")
	params.Set("chart", "mostPopular")
	params.Set("regionCode", "US")
	params.Set("maxResults", fmt.Sprintf("%d", limit))
	params.Set("key", y.apiKey)

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch trending videos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				Title       string    `json:"title"`
				ChannelID   string    `json:"channelId"`
				PublishedAt time.Time `json:"publishedAt"`
			} `json:"snippet"`
			Statistics struct {
				ViewCount string `json:"viewCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	videos := make([]*Video, 0, len(result.Items))
	for _, item := range result.Items {
		viewCount := parseInt64(item.Statistics.ViewCount)
		videos = append(videos, &Video{
			ID:          item.ID,
			Title:       item.Snippet.Title,
			ViewCount:   viewCount,
			ChannelID:   item.Snippet.ChannelID,
			PublishedAt: item.Snippet.PublishedAt,
		})
	}

	return videos, nil
}

// GetChannelStats fetches channel statistics
func (y *YouTubePlatform) GetChannelStats(ctx context.Context) (*ChannelStats, error) {
	endpoint := fmt.Sprintf("%s/channels", y.baseURL)

	params := url.Values{}
	params.Set("part", "statistics")
	params.Set("id", y.channelID)
	params.Set("key", y.apiKey)

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			Statistics struct {
				SubscriberCount string `json:"subscriberCount"`
				ViewCount       string `json:"viewCount"`
				VideoCount      string `json:"videoCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("channel not found")
	}

	stats := result.Items[0].Statistics
	return &ChannelStats{
		SubscriberCount: parseInt64(stats.SubscriberCount),
		ViewCount:       parseInt64(stats.ViewCount),
		VideoCount:      parseInt64(stats.VideoCount),
	}, nil
}

// GetRecentComments fetches recent comments from channel videos
func (y *YouTubePlatform) GetRecentComments(ctx context.Context, limit int) ([]*Comment, error) {
	// First, get recent videos from the channel
	videos, err := y.getChannelVideos(ctx, 5)
	if err != nil {
		return nil, err
	}

	if len(videos) == 0 {
		return []*Comment{}, nil
	}

	// Get comments from the most recent video
	videoID := videos[0].ID
	return y.getVideoComments(ctx, videoID, limit)
}

// getChannelVideos fetches recent videos from the channel
func (y *YouTubePlatform) getChannelVideos(ctx context.Context, limit int) ([]*Video, error) {
	endpoint := fmt.Sprintf("%s/search", y.baseURL)

	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("channelId", y.channelID)
	params.Set("order", "date")
	params.Set("type", "video")
	params.Set("maxResults", fmt.Sprintf("%d", limit))
	params.Set("key", y.apiKey)

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel videos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title       string    `json:"title"`
				PublishedAt time.Time `json:"publishedAt"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	videos := make([]*Video, 0, len(result.Items))
	for _, item := range result.Items {
		videos = append(videos, &Video{
			ID:          item.ID.VideoID,
			Title:       item.Snippet.Title,
			PublishedAt: item.Snippet.PublishedAt,
		})
	}

	return videos, nil
}

// getVideoComments fetches comments for a specific video
func (y *YouTubePlatform) getVideoComments(ctx context.Context, videoID string, limit int) ([]*Comment, error) {
	endpoint := fmt.Sprintf("%s/commentThreads", y.baseURL)

	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("videoId", videoID)
	params.Set("order", "time")
	params.Set("maxResults", fmt.Sprintf("%d", limit))
	params.Set("key", y.apiKey)

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			ID      string `json:"id"`
			Snippet struct {
				TopLevelComment struct {
					ID      string `json:"id"`
					Snippet struct {
						AuthorDisplayName string    `json:"authorDisplayName"`
						TextDisplay       string    `json:"textDisplay"`
						PublishedAt       time.Time `json:"publishedAt"`
					} `json:"snippet"`
				} `json:"topLevelComment"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	comments := make([]*Comment, 0, len(result.Items))
	for _, item := range result.Items {
		comment := item.Snippet.TopLevelComment
		comments = append(comments, &Comment{
			ID:        comment.ID,
			VideoID:   videoID,
			Author:    comment.Snippet.AuthorDisplayName,
			Text:      comment.Snippet.TextDisplay,
			CreatedAt: comment.Snippet.PublishedAt,
		})
	}

	return comments, nil
}

// Helper function to parse int64 from string
func parseInt64(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}
