package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/cwat-code/cwat-go-client/internal/provider"
)

func BashTool() provider.Tool {
	return provider.Tool{
		Name:        "bash",
		Description: "Execute a bash/powershell command. Provides access to a system shell.",
		InputSchema: json.RawMessage(`{"type": "object", "properties": {"command": {"type": "string"}}, "required": ["command"]}`),
	}
}

type BashArgs struct {
	Command string `json:"command"`
}

func ExecuteBash(ctx context.Context, argsRaw json.RawMessage) string {
	var args BashArgs
	if err := json.Unmarshal(argsRaw, &args); err != nil {
		return fmt.Sprintf("Error parsing args: %v", err)
	}

	execCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "powershell", "-NoProfile", "-NonInteractive", "-Command", args.Command)
	} else {
		cmd = exec.CommandContext(execCtx, "bash", "-c", args.Command)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			return fmt.Sprintf("Command timed out after 30 seconds. Output so far:\n%s", string(out))
		}
		return fmt.Sprintf("Error executing command:\n%v\nOutput:\n%s", err, string(out))
	}
	res := string(out)
	if res == "" {
		return "Command executed successfully with no output."
	}
	return res
}
