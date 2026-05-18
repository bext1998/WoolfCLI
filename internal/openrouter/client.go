package openrouter

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	MaxRetries int
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream"`
}

func (c *Client) StreamChat(ctx context.Context, req ChatRequest) (<-chan StreamEvent, error) {
	if strings.TrimSpace(c.APIKey) == "" {
		return nil, APIError{Code: "CFG-001", Message: "OpenRouter API key is required"}
	}
	if c.BaseURL == "" {
		c.BaseURL = "https://openrouter.ai/api/v1"
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 120 * time.Second}
	}
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.doWithRetry(ctx, body)
	if err != nil {
		return nil, err
	}

	events := make(chan StreamEvent)
	go func() {
		defer close(events)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" {
				events <- StreamEvent{Done: true}
				return
			}
			event, err := decodeStreamPayload(payload)
			if err != nil {
				events <- StreamEvent{Error: err}
				return
			}
			events <- event
		}
		if err := scanner.Err(); err != nil {
			events <- StreamEvent{Error: err}
		}
	}()
	return events, nil
}

func (c *Client) doWithRetry(ctx context.Context, body []byte) (*http.Response, error) {
	attempts := c.MaxRetries + 1
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := c.do(ctx, body)
		if err != nil {
			lastErr = err
		} else if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			lastErr = apiErrorFromResponse(resp)
			drainAndClose(resp.Body)
		} else if resp.StatusCode >= 400 {
			err := apiErrorFromResponse(resp)
			drainAndClose(resp.Body)
			return nil, err
		} else {
			return resp, nil
		}
		if attempt+1 < attempts {
			time.Sleep(retryDelay(resp, attempt))
		}
	}
	return nil, lastErr
}

func (c *Client) do(ctx context.Context, body []byte) (*http.Response, error) {
	url := strings.TrimRight(c.BaseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	httpReq.Header.Set("HTTP-Referer", "https://github.com/woolf-cli")
	httpReq.Header.Set("X-Title", "Woolf")
	return c.HTTPClient.Do(httpReq)
}

func decodeStreamPayload(payload string) (StreamEvent, error) {
	var raw struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
			FinishReason *string `json:"finish_reason"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		return StreamEvent{}, err
	}
	if raw.Error != nil {
		code := raw.Error.Code
		if code == "" {
			code = "API-ERROR"
		}
		return StreamEvent{}, APIError{Code: code, Message: raw.Error.Message}
	}
	event := StreamEvent{}
	if len(raw.Choices) > 0 {
		event.Content = raw.Choices[0].Delta.Content
		event.Done = raw.Choices[0].FinishReason != nil
	}
	if raw.Usage != nil {
		event.Usage = &Usage{
			PromptTokens:     raw.Usage.PromptTokens,
			CompletionTokens: raw.Usage.CompletionTokens,
			TotalTokens:      raw.Usage.TotalTokens,
		}
	}
	return event, nil
}

func apiErrorFromResponse(resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	msg := strings.TrimSpace(string(body))
	if msg == "" {
		msg = resp.Status
	}
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return APIError{Code: "API-001", Message: msg}
	case http.StatusPaymentRequired:
		return APIError{Code: "API-002", Message: msg}
	case http.StatusTooManyRequests:
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			msg = fmt.Sprintf("%s retry_after=%s", msg, retryAfter)
		}
		return APIError{Code: "API-003", Message: msg}
	case http.StatusNotFound:
		return APIError{Code: "API-004", Message: msg}
	default:
		if resp.StatusCode >= 500 {
			return APIError{Code: "API-005", Message: msg}
		}
		return APIError{Code: "API-ERROR", Message: msg}
	}
}

func retryDelay(resp *http.Response, attempt int) time.Duration {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		if value := strings.TrimSpace(resp.Header.Get("Retry-After")); value != "" {
			if seconds, err := time.ParseDuration(value + "s"); err == nil {
				return seconds
			}
			if when, err := http.ParseTime(value); err == nil {
				delay := time.Until(when)
				if delay > 0 {
					return delay
				}
				return 0
			}
		}
	}
	return time.Duration(1<<attempt) * time.Second
}

func drainAndClose(body io.ReadCloser) {
	io.Copy(io.Discard, io.LimitReader(body, 4096))
	body.Close()
}
