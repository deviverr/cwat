package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AnthropicClient supports Anthropic Messages API.
type AnthropicClient struct {
	httpClient       *http.Client
	endpoint         string
	apiKey           string
	anthropicVersion string
}

// NewAnthropicClient builds a new Anthropic API client.
func NewAnthropicClient(baseURL string, apiKey string, anthropicVersion string) *AnthropicClient {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = "https://api.anthropic.com/v1/messages"
	}
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/v1") {
		base += "/messages"
	} else if !strings.HasSuffix(base, "/messages") {
		base += "/v1/messages"
	}

	version := strings.TrimSpace(anthropicVersion)
	if version == "" {
		version = "2023-06-01"
	}

	return &AnthropicClient{
		httpClient:       &http.Client{Timeout: 90 * time.Second},
		endpoint:         base,
		apiKey:           strings.TrimSpace(apiKey),
		anthropicVersion: version,
	}
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	MaxTokens   int                `json:"max_tokens"`
	Temperature float64            `json:"temperature,omitempty"`
}

type anthropicMessage struct {
	Role    string               `json:"role"`
	Content []anthropicTextBlock `json:"content"`
}

type anthropicTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Model   string `json:"model"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Chat sends one request to Anthropic Messages API.
func (c *AnthropicClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if strings.TrimSpace(req.Model) == "" {
		return ChatResponse{}, fmt.Errorf("model is required")
	}
	if len(req.Messages) == 0 {
		return ChatResponse{}, fmt.Errorf("at least one message is required")
	}

	systemParts := make([]string, 0)
	messages := make([]anthropicMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		switch m.Role {
		case RoleSystem:
			if strings.TrimSpace(m.Content) != "" {
				systemParts = append(systemParts, m.Content)
			}
		case RoleUser, RoleAssistant:
			messages = append(messages, anthropicMessage{
				Role: string(m.Role),
				Content: []anthropicTextBlock{
					{Type: "text", Text: m.Content},
				},
			})
		default:
			return ChatResponse{}, fmt.Errorf("unsupported role for anthropic: %s", m.Role)
		}
	}
	if len(messages) == 0 {
		return ChatResponse{}, fmt.Errorf("anthropic request has no non-system messages")
	}

	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	payload := anthropicRequest{
		Model:       req.Model,
		Messages:    messages,
		System:      strings.TrimSpace(strings.Join(systemParts, "\n\n")),
		MaxTokens:   maxTokens,
		Temperature: req.Temperature,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(b))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", c.anthropicVersion)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := readLimitedBody(httpResp.Body, maxProviderResponseBytes)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("read response: %w", err)
	}

	var parsed anthropicResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
			if snippet := truncateForError(body, 400); snippet != "" {
				return ChatResponse{}, fmt.Errorf("request failed (status %d): %s", httpResp.StatusCode, snippet)
			}
			return ChatResponse{}, fmt.Errorf("request failed (status %d)", httpResp.StatusCode)
		}
		return ChatResponse{}, fmt.Errorf("parse response: %w", err)
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		em := ""
		if parsed.Error != nil && strings.TrimSpace(parsed.Error.Message) != "" {
			em = parsed.Error.Message
		}
		if em == "" {
			em = truncateForError(body, 400)
		}
		if em == "" {
			em = "request failed"
		}
		return ChatResponse{}, fmt.Errorf("%s (status %d)", em, httpResp.StatusCode)
	}
	if len(parsed.Content) == 0 {
		return ChatResponse{}, fmt.Errorf("provider returned empty content")
	}

	parts := make([]string, 0, len(parsed.Content))
	for _, block := range parsed.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			parts = append(parts, block.Text)
		}
	}
	text := strings.TrimSpace(strings.Join(parts, "\n"))
	if text == "" {
		return ChatResponse{}, fmt.Errorf("provider returned no text blocks")
	}

	model := parsed.Model
	if strings.TrimSpace(model) == "" {
		model = req.Model
	}
	return ChatResponse{Provider: "anthropic", Model: model, Text: text}, nil
}
