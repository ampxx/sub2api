package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultTimeout is the default HTTP client timeout for fetching subscriptions
	DefaultTimeout = 15 * time.Second
	// MaxResponseSize limits the subscription response body to 10MB
	MaxResponseSize = 10 * 1024 * 1024
)

// SubscriptionHandler holds dependencies for subscription-related handlers
type SubscriptionHandler struct {
	Client *http.Client
}

// NewSubscriptionHandler creates a new SubscriptionHandler with a configured HTTP client
func NewSubscriptionHandler(timeout time.Duration) *SubscriptionHandler {
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	return &SubscriptionHandler{
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchSubscription fetches a remote subscription URL and returns its raw content.
// Query param: url — the subscription URL to fetch
func (h *SubscriptionHandler) FetchSubscription(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing required query parameter: url",
		})
		return
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid subscription URL; must be http or https",
		})
		return
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, rawURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to build request: %v", err),
		})
		return
	}

	// Forward a generic User-Agent to avoid blocks
	req.Header.Set("User-Agent", "sub2api/1.0 (+https://github.com/sub2api/sub2api)")

	resp, err := h.Client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": fmt.Sprintf("failed to fetch subscription: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadGateway, gin.H{
			"error": fmt.Sprintf("upstream returned status %d", resp.StatusCode),
		})
		return
	}

	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to read response body: %v", err),
		})
		return
	}

	// Pass through content-type if present, default to text/plain
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" || strings.Contains(contentType, "application/octet-stream") {
		contentType = "text/plain; charset=utf-8"
	}

	c.Data(http.StatusOK, contentType, body)
}
