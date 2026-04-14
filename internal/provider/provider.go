package provider

import "context"

// Role is the chat message author role.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message is a normalized provider-agnostic chat message.
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"` // For tool outputs
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"` // For assistant calls
}

// ChatRequest is a normalized chat completion request.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// ChatResponse is a normalized chat completion response.
type ChatResponse struct {
	Provider  string     `json:"provider"`
	Model     string     `json:"model"`
	Text      string     `json:"text"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// Client is the minimal provider interface used by cwat.
type Client interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
