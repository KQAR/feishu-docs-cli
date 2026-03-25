package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/KQAR/feishu-docs-cli/internal/config"
)

const (
	rateLimitMinInterval = 300 * time.Millisecond
	rateLimitMaxRetries  = 3
	rateLimitBaseDelay   = time.Second
)

// New 根据配置创建飞书 SDK 客户端
func New(cfg *config.Config) *lark.Client {
	httpClient := &rateLimitClient{
		base:       http.DefaultClient,
		minDelay:   rateLimitMinInterval,
		maxRetries: rateLimitMaxRetries,
		baseDelay:  rateLimitBaseDelay,
	}
	return lark.NewClient(cfg.AppID, cfg.AppSecret, lark.WithHttpClient(httpClient))
}

// rateLimitClient wraps an http.Client with automatic throttling and 429 retry.
type rateLimitClient struct {
	base       *http.Client
	minDelay   time.Duration
	maxRetries int
	baseDelay  time.Duration
	mu         sync.Mutex
	lastCall   time.Time
}

func (c *rateLimitClient) Do(req *http.Request) (*http.Response, error) {
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("reading request body for retry: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		c.mu.Lock()
		elapsed := time.Since(c.lastCall)
		if elapsed < c.minDelay {
			time.Sleep(c.minDelay - elapsed)
		}
		c.lastCall = time.Now()
		c.mu.Unlock()

		resp, err := c.base.Do(req)
		if err != nil {
			return resp, err
		}
		if resp.StatusCode != http.StatusTooManyRequests || attempt == c.maxRetries {
			return resp, nil
		}

		resp.Body.Close()
		delay := c.baseDelay * time.Duration(1<<uint(attempt))
		fmt.Fprintf(os.Stderr, "⚠ API 限流 (429)，%v 后重试 (%d/%d)...\n", delay, attempt+1, c.maxRetries)
		time.Sleep(delay)

		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	return nil, fmt.Errorf("rate limit retries exhausted")
}
