package openrouter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStreamChatParsesSSE(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hel\"}}]}\n\n"))
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}],\"usage\":{\"prompt_tokens\":2,\"completion_tokens\":3,\"total_tokens\":5}}\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTPClient: server.Client()}
	stream, err := client.StreamChat(context.Background(), ChatRequest{Model: "m", Messages: []ChatMessage{{Role: "user", Content: "hi"}}})
	if err != nil {
		t.Fatalf("StreamChat() error = %v", err)
	}
	var content string
	var usage *Usage
	for event := range stream {
		content += event.Content
		if event.Usage != nil {
			usage = event.Usage
		}
	}
	if content != "hello" {
		t.Fatalf("content = %q", content)
	}
	if usage == nil || usage.TotalTokens != 5 {
		t.Fatalf("usage = %#v", usage)
	}
}

func TestStreamChatMapsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "no credit", http.StatusPaymentRequired)
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "key", HTTPClient: server.Client()}
	_, err := client.StreamChat(context.Background(), ChatRequest{Model: "m"})
	if err == nil || !strings.Contains(err.Error(), "API-002") {
		t.Fatalf("StreamChat() error = %v, want API-002", err)
	}
}

func TestStreamChatRequiresAPIKey(t *testing.T) {
	client := &Client{BaseURL: "https://example.test"}
	_, err := client.StreamChat(context.Background(), ChatRequest{Model: "m"})
	if err == nil || !strings.Contains(err.Error(), "CFG-001") {
		t.Fatalf("StreamChat() error = %v, want CFG-001", err)
	}
}

func TestStreamChatMapsStatusCodes(t *testing.T) {
	tests := []struct {
		name   string
		status int
		code   string
	}{
		{name: "auth", status: http.StatusUnauthorized, code: "API-001"},
		{name: "model", status: http.StatusNotFound, code: "API-004"},
		{name: "server", status: http.StatusBadGateway, code: "API-005"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, tt.name, tt.status)
			}))
			defer server.Close()

			client := &Client{
				BaseURL:    server.URL,
				APIKey:     "key",
				HTTPClient: server.Client(),
				MaxRetries: 1,
				RetrySleep: func(time.Duration) {},
			}
			_, err := client.StreamChat(context.Background(), ChatRequest{Model: "m"})
			if err == nil || !strings.Contains(err.Error(), tt.code) {
				t.Fatalf("StreamChat() error = %v, want %s", err, tt.code)
			}
		})
	}
}

func TestStreamChatRetriesRateLimitWithRetryAfter(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "2")
			http.Error(w, "slow down", http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	var slept []time.Duration
	client := &Client{
		BaseURL:    server.URL,
		APIKey:     "key",
		HTTPClient: server.Client(),
		MaxRetries: 1,
		RetrySleep: func(delay time.Duration) {
			slept = append(slept, delay)
		},
	}
	stream, err := client.StreamChat(context.Background(), ChatRequest{Model: "m"})
	if err != nil {
		t.Fatalf("StreamChat() error = %v", err)
	}
	for range stream {
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
	if len(slept) != 1 || slept[0] != 2*time.Second {
		t.Fatalf("slept = %#v, want 2s", slept)
	}
}

func TestStreamChatDefaultRetriesServerErrors(t *testing.T) {
	var attempts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		http.Error(w, "temporary", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		APIKey:     "key",
		HTTPClient: server.Client(),
		RetrySleep: func(time.Duration) {},
	}
	_, err := client.StreamChat(context.Background(), ChatRequest{Model: "m"})
	if err == nil || !strings.Contains(err.Error(), "API-005") {
		t.Fatalf("StreamChat() error = %v, want API-005", err)
	}
	if attempts != 4 {
		t.Fatalf("attempts = %d, want initial request plus 3 retries", attempts)
	}
}
