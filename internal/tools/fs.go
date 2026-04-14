package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cwat-code/cwat-go-client/internal/provider"
)

func ReadFileTool() provider.Tool {
	return provider.Tool{
		Name:        "read_file",
		Description: "Read the contents of a file.",
		InputSchema: json.RawMessage(`{"type": "object", "properties": {"filepath": {"type": "string"}}, "required": ["filepath"]}`),
	}
}

type ReadFileArgs struct {
	Filepath string `json:"filepath"`
}

func ExecuteReadFile(ctx context.Context, argsRaw json.RawMessage) string {
	var args ReadFileArgs
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return fmt.Sprintf("Error parsing args: %v", err)
	}

	content, err := os.ReadFile(args.Filepath)
	if err != nil {
		return fmt.Sprintf("Error reading file %s: %v", args.Filepath, err)
	}
	return string(content)
}

func WriteFileTool() provider.Tool {
	return provider.Tool{
		Name:        "write_file",
		Description: "Write content to a file. Overwrites existing file.",
		InputSchema: json.RawMessage(`{"type": "object", "properties": {"filepath": {"type": "string"}, "content": {"type": "string"}}, "required": ["filepath", "content"]}`),
	}
}

type WriteFileArgs struct {
	Filepath string `json:"filepath"`
	Content  string `json:"content"`
}

func ExecuteWriteFile(ctx context.Context, argsRaw json.RawMessage) string {
	var args WriteFileArgs
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return fmt.Sprintf("Error parsing args: %v", err)
	}

	if err := os.WriteFile(args.Filepath, []byte(args.Content), 0644); err != nil {
		return fmt.Sprintf("Error writing file %s: %v", args.Filepath, err)
	}
	return fmt.Sprintf("Successfully wrote to %s", args.Filepath)
}
