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

type OpenAIClient struct {
	httpClient   *http.Client
	baseURL      string
	apiKey       string
	providerName string
}

func NewOpenAIClient(baseURL string, apiKey string, providerName string) *OpenAIClient {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	provider := strings.TrimSpace(providerName)
	if provider == "" {
		provider = "openai"
	}
	return &OpenAIClient{
		httpClient:   &http.Client{Timeout: 90 * time.Second},
		baseURL:      strings.TrimRight(base, "/"),
		apiKey:       strings.TrimSpace(apiKey),
		providerName: provider,
	}
}

func (c *OpenAIClient) endpoint() string {
	if strings.HasSuffix(c.baseURL, "/chat/completions") {
		return c.baseURL
	}
	return c.baseURL + "/chat/completions"
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	Tools       []openAITool    `json:"tools,omitempty"`
	ToolChoice  interface{}     `json:"tool_choice,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
}

type openAITool struct {
	Type     string             `json:"type"`
	Function openAIToolFunction `json:"function"`
}

type openAIToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content"`
	Name       string           `json:"name,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAIResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content   json.RawMessage  `json:"content"`
			ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *OpenAIClient) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if strings.TrimSpace(req.Model) == "" {
		return ChatResponse{}, fmt.Errorf("model is required")
	}
	if len(req.Messages) == 0 {
		return ChatResponse{}, fmt.Errorf("at least one message is required")
	}

	messages := make([]openAIMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		openMsg := openAIMessage{Role: string(m.Role), Content: m.Content}
		if m.Name != "" {
			openMsg.Name = m.Name
		}
		if m.ToolCallID != "" {
			openMsg.ToolCallID = m.ToolCallID
		}
		for _, tc := range m.ToolCalls {
			openMsg.ToolCalls = append(openMsg.ToolCalls, openAIToolCall{
				ID:   tc.ID,
				Type: tc.Type,
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      tc.Name,
					Arguments: string(tc.Arguments),
				},
			})
		}
		messages = append(messages, openMsg)
	}

	var openTools []openAITool
	for _, t := range req.Tools {
		openTools = append(openTools, openAITool{
			Type: "function",
			Function: openAIToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}

	payload := openAIRequest{
		Model:    req.Model,
		Messages: messages,
		Tools:    openTools,
		// ToolChoice omitted to avoid litellm errors with Qwen
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint(), bytes.NewReader(b))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("send request: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := readLimitedBody(httpResp.Body, maxProviderResponseBytes)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("read response: %w", err)
	}

	var parsed openAIResponse
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
	if len(parsed.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("provider returned no choices")
	}

	text := extractOpenAIContent(parsed.Choices[0].Message.Content)
	if strings.TrimSpace(text) == "" && len(parsed.Choices[0].Message.ToolCalls) == 0 {
		return ChatResponse{}, fmt.Errorf("provider returned empty response")
	}

	var resToolCalls []ToolCall
	for _, tc := range parsed.Choices[0].Message.ToolCalls {
		resToolCalls = append(resToolCalls, ToolCall{
			ID:        tc.ID,
			Type:      tc.Type,
			Name:      tc.Function.Name,
			Arguments: json.RawMessage(tc.Function.Arguments),
		})
	}

	model := parsed.Model
	if strings.TrimSpace(model) == "" {
		model = req.Model
	}
	return ChatResponse{Provider: c.providerName, Model: model, Text: text, ToolCalls: resToolCalls}, nil
}

func extractOpenAIContent(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var plain string
	if err := json.Unmarshal(raw, &plain); err == nil {
		return strings.TrimSpace(plain)
	}
	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &blocks); err == nil {
		parts := make([]string, 0, len(blocks))
		for _, block := range blocks {
			if (block.Type == "" || block.Type == "text") && strings.TrimSpace(block.Text) != "" {
				parts = append(parts, block.Text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	}
	return ""
}
