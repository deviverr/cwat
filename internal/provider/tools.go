package provider

import "encoding/json"

// Tool represents a function the model can invoke.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"` // Contains JSON Schema object
}

// ToolCall represents a specific tool invocation requested by the model.
type ToolCall struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"` // "function"
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"` // JSON args
}

// ToolMessage represents the result of executing a tool call.
type ToolMessage struct {
	ToolCallID string `json:"tool_call_id"`
	Name       string `json:"name,omitempty"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error,omitempty"`
}
