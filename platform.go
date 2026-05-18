package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

// ChannelStats represents channel analytics with recent performance
type ChannelStats struct {
	SubscriberCount int64
	RecentVideos    []*RecentVideo
}

// RecentVideo represents a recent video's performance
type RecentVideo struct {
	Title        string
	ViewCount    int64
	LikeCount    int64
	CommentCount int64
	PublishedAt  time.Time
	AgeInDays    int
}

// YouTubePlatform handles all YouTube API operations
type YouTubePlatform struct {
	apiKey         string
	channelID      string
	client         *http.Client
	baseURL        string
	trendCategory  string // Override category for trending (empty = auto-detect)
	trendRegion    string // Region code for trending videos
}

// NewYouTubePlatform creates a new YouTube platform provider
func NewYouTubePlatform(apiKey, channelID, trendCategory, trendRegion string) *YouTubePlatform {
	if trendRegion == "" {
		trendRegion = "US"
	}
	return &YouTubePlatform{
		apiKey:        apiKey,
		channelID:     channelID,
		trendCategory: trendCategory,
		trendRegion:   trendRegion,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://www.googleapis.com/youtube/v3",
	}
}

// GetTrendingVideos fetches trending videos based on configured category and region
func (y *YouTubePlatform) GetTrendingVideos(ctx context.Context, limit int) ([]*Video, error) {
	// Use configured category, or auto-detect from channel if not set
	categoryID := y.trendCategory
	if categoryID == "" {
		var err error
		categoryID, err = y.getChannelCategory(ctx)
		if err != nil {
			// Fallback to generic trending if we can't get category
			categoryID = ""
		}
	}

	endpoint := fmt.Sprintf("%s/videos", y.baseURL)

	params := url.Values{}
	params.Set("part", "snippet,statistics")
	params.Set("chart", "mostPopular")
	params.Set("regionCode", y.trendRegion)
	params.Set("maxResults", fmt.Sprintf("%d", limit*2)) // Get more to filter
	if categoryID != "" {
		params.Set("videoCategoryId", categoryID)
	}
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

	videos := make([]*Video, 0, limit)
	for _, item := range result.Items {
		// Skip your own videos
		if item.Snippet.ChannelID == y.channelID {
			continue
		}

		viewCount := parseInt64(item.Statistics.ViewCount)
		videos = append(videos, &Video{
			ID:          item.ID,
			Title:       item.Snippet.Title,
			ViewCount:   viewCount,
			ChannelID:   item.Snippet.ChannelID,
			PublishedAt: item.Snippet.PublishedAt,
		})

		if len(videos) >= limit {
			break
		}
	}

	return videos, nil
}

// getChannelCategory fetches the primary category of the channel's videos
func (y *YouTubePlatform) getChannelCategory(ctx context.Context) (string, error) {
	// Get a recent video from the channel
	videos, err := y.getChannelVideos(ctx, 1)
	if err != nil || len(videos) == 0 {
		return "", err
	}

	// Get the video details to find its category
	endpoint := fmt.Sprintf("%s/videos", y.baseURL)
	params := url.Values{}
	params.Set("part", "snippet")
	params.Set("id", videos[0].ID)
	params.Set("key", y.apiKey)

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := y.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get video category")
	}

	var result struct {
		Items []struct {
			Snippet struct {
				CategoryID string `json:"categoryId"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Items) > 0 {
		return result.Items[0].Snippet.CategoryID, nil
	}

	return "", fmt.Errorf("no category found")
}

// GetChannelStats fetches channel statistics with recent video performance
func (y *YouTubePlatform) GetChannelStats(ctx context.Context) (*ChannelStats, error) {
	// Get subscriber count
	subscriberCount, err := y.getSubscriberCount(ctx)
	if err != nil {
		return nil, err
	}

	// Get recent videos
	videos, err := y.getChannelVideos(ctx, 5)
	if err != nil {
		return nil, err
	}

	if len(videos) == 0 {
		return &ChannelStats{
			SubscriberCount: subscriberCount,
			RecentVideos:    []*RecentVideo{},
		}, nil
	}

	// Get detailed stats for each video
	recentVideos, err := y.getVideoStats(ctx, videos)
	if err != nil {
		return nil, err
	}

	return &ChannelStats{
		SubscriberCount: subscriberCount,
		RecentVideos:    recentVideos,
	}, nil
}

// getSubscriberCount fetches just the subscriber count
func (y *YouTubePlatform) getSubscriberCount(ctx context.Context) (int64, error) {
	endpoint := fmt.Sprintf("%s/channels", y.baseURL)

	params := url.Values{}
	params.Set("part", "statistics")
	params.Set("id", y.channelID)
	params.Set("key", y.apiKey)

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := y.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch channel stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("YouTube API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			Statistics struct {
				SubscriberCount string `json:"subscriberCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(result.Items) == 0 {
		return 0, fmt.Errorf("channel not found")
	}

	return parseInt64(result.Items[0].Statistics.SubscriberCount), nil
}

// getVideoStats fetches detailed statistics for multiple videos
func (y *YouTubePlatform) getVideoStats(ctx context.Context, videos []*Video) ([]*RecentVideo, error) {
	// Build video IDs string
	var videoIDs []string
	for _, v := range videos {
		videoIDs = append(videoIDs, v.ID)
	}

	endpoint := fmt.Sprintf("%s/videos", y.baseURL)

	params := url.Values{}
	params.Set("part", "snippet,statistics")
	params.Set("id", strings.Join(videoIDs, ","))
	params.Set("key", y.apiKey)

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := y.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video stats: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YouTube API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Items []struct {
			Snippet struct {
				Title       string    `json:"title"`
				PublishedAt time.Time `json:"publishedAt"`
			} `json:"snippet"`
			Statistics struct {
				ViewCount    string `json:"viewCount"`
				LikeCount    string `json:"likeCount"`
				CommentCount string `json:"commentCount"`
			} `json:"statistics"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	recentVideos := make([]*RecentVideo, 0, len(result.Items))
	now := time.Now()

	for _, item := range result.Items {
		ageInDays := int(now.Sub(item.Snippet.PublishedAt).Hours() / 24)
		recentVideos = append(recentVideos, &RecentVideo{
			Title:        item.Snippet.Title,
			ViewCount:    parseInt64(item.Statistics.ViewCount),
			LikeCount:    parseInt64(item.Statistics.LikeCount),
			CommentCount: parseInt64(item.Statistics.CommentCount),
			PublishedAt:  item.Snippet.PublishedAt,
			AgeInDays:    ageInDays,
		})
	}

	return recentVideos, nil
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
