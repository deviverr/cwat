package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/charmbracelet/glamour"
	"github.com/cwat-code/cwat-go-client/internal/provider"
	"github.com/cwat-code/cwat-go-client/internal/tools"
	"github.com/cwat-code/cwat-go-client/internal/ui"
)

// SessionOptions controls chat runtime behavior.
type SessionOptions struct {
	Model          string
	System         string
	MaxTokens      int
	Temperature    float64
	RequestTimeout time.Duration
}

// RunSinglePrompt executes one prompt and returns the assistant text.
func RunSinglePrompt(ctx context.Context, client provider.Client, prompt string, opts SessionOptions) (string, error) {
	messages := make([]provider.Message, 0, 2)
	if strings.TrimSpace(opts.System) != "" {
		messages = append(messages, provider.Message{Role: provider.RoleSystem, Content: opts.System})
	}
	messages = append(messages, provider.Message{Role: provider.RoleUser, Content: prompt})

	reqCtx := ctx
	if opts.RequestTimeout > 0 {
		var cancel context.CancelFunc
		reqCtx, cancel = context.WithTimeout(ctx, opts.RequestTimeout)
		defer cancel()
	}

	resp, err := client.Chat(reqCtx, provider.ChatRequest{
		Model:       opts.Model,
		Messages:    messages,
		MaxTokens:   opts.MaxTokens,
		Temperature: opts.Temperature,
	})
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}

func getHistoryPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "/tmp/cwat_history.tmp"
	}
	cwatDir := filepath.Join(configDir, "cwat")
	_ = os.MkdirAll(cwatDir, 0755)
	return filepath.Join(cwatDir, "history.tmp")
}

// RunInteractive starts a local REPL that keeps conversation history.
func RunInteractive(ctx context.Context, client provider.Client, opts SessionOptions) error {
	history := make([]provider.Message, 0, 16)
	if strings.TrimSpace(opts.System) != "" {
		history = append(history, provider.Message{Role: provider.RoleSystem, Content: opts.System})
	}

	ui.PrintDashboard(opts.Model)

	completer := func(d prompt.Document) []prompt.Suggest {
		s := []prompt.Suggest{
			{Text: "/help", Description: "Show commands"},
			{Text: "/bug", Description: "Report a bug"},
			{Text: "/clear", Description: "Clear conversation history"},
			{Text: "/compact", Description: "Make chat compact"},
			{Text: "/config", Description: "Show configuration"},
			{Text: "/cost", Description: "Show estimated cost"},
			{Text: "/doctor", Description: "Run diagnostic checks"},
			{Text: "/init", Description: "Initialize cwat"},
			{Text: "/login", Description: "Login to sync"},
			{Text: "/logout", Description: "Logout from sync"},
			{Text: "/pr-comments", Description: "Fetch PR comments"},
			{Text: "/resume", Description: "Resume a session"},
			{Text: "/terminal", Description: "Run a terminal command manually"},
			{Text: "/theme", Description: "Change UI theme"},
			{Text: "/exit", Description: "Quit cwat"},
		}
		if len(d.TextBeforeCursor()) > 0 && d.TextBeforeCursor()[0] == '/' {
			return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
		}
		return []prompt.Suggest{}
	}

	for {
		ui.PrintPrompt("")

		line := prompt.Input("❯ ", completer,
			prompt.OptionTitle("cwat"),
			prompt.OptionHistory([]string{}),
			prompt.OptionPrefixTextColor(prompt.Cyan),
			prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
			prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
			prompt.OptionSuggestionBGColor(prompt.DarkGray),
		)

		if line == "exit" || line == "/exit" {
			return nil
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		switch line {
		case "/help":
			fmt.Println("  /help         show commands")
			fmt.Println("  /bug          report a bug")
			fmt.Println("  /clear        clear conversation history")
			fmt.Println("  /compact      make chat compact")
			fmt.Println("  /config       show configuration")
			fmt.Println("  /cost         show estimated cost")
			fmt.Println("  /doctor       run diagnostic checks")
			fmt.Println("  /init         initialize cwat")
			fmt.Println("  /login        login to sync")
			fmt.Println("  /logout       logout from sync")
			fmt.Println("  /pr-comments  fetch PR comments")
			fmt.Println("  /resume       resume a session")
			fmt.Println("  /terminal     run a terminal command manually")
			fmt.Println("  /theme        change UI theme")
			fmt.Println("  /exit         quit cwat")
			continue
		case "/cost":
			fmt.Println(ui.UiMuted("Session Cost: $0.0000 (tracking TBD)"))
			continue
		case "/theme":
			fmt.Println(ui.UiSuccess("Theme changed to standard cyan/ocean-blue (cwat default)."))
			continue
		case "/exit":
			return nil
		case "/clear":
			if strings.TrimSpace(opts.System) != "" {
				history = []provider.Message{{Role: provider.RoleSystem, Content: opts.System}}
			} else {
				history = history[:0]
			}
			fmt.Println(ui.UiSuccess("History cleared."))
			continue
		case "/bug", "/compact", "/config", "/doctor", "/init", "/login", "/logout", "/pr-comments", "/resume", "/terminal":
			fmt.Println(ui.UiMuted("Mocked: Command not fully implemented in Go rewrite yet."))
			continue
		}

		history = append(history, provider.Message{Role: provider.RoleUser, Content: line})

		for {
			reqCtx := ctx
			cancel := func() {}
			if opts.RequestTimeout > 0 {
				reqCtx, cancel = context.WithTimeout(ctx, opts.RequestTimeout)
			}

			spinner := ui.StartSpinner("Working...")

			resp, err := client.Chat(reqCtx, provider.ChatRequest{
				Model:       opts.Model,
				Messages:    history,
				Tools:       []provider.Tool{tools.BashTool(), tools.ReadFileTool(), tools.WriteFileTool()},
				MaxTokens:   opts.MaxTokens,
				Temperature: opts.Temperature,
			})
			cancel()
			spinner.Stop()

			if err != nil {
				fmt.Println(ui.UiError(fmt.Sprintf("error: %v", err)))
				// Drop user line if request failed so broken turns do not pollute history.
				for len(history) > 0 && history[len(history)-1].Role != provider.RoleUser {
					history = history[:len(history)-1] // Roll back tools/assistant
				}
				if len(history) > 0 {
					history = history[:len(history)-1] // Drop user message
				}
				break
			}

			// Add assistant message (which might contain tool calls)
			history = append(history, provider.Message{
				Role:      provider.RoleAssistant,
				Content:   resp.Text,
				ToolCalls: resp.ToolCalls,
			})

			if resp.Text != "" {
				if r, err := glamour.NewTermRenderer(
					glamour.WithAutoStyle(),
					glamour.WithWordWrap(ui.GetTermWidth()),
				); err == nil {
					out, _ := r.Render(strings.TrimSpace(resp.Text))
					fmt.Print(out)
				} else {
					fmt.Printf("\n%s\n\n", strings.TrimSpace(resp.Text))
				}
			}

			if len(resp.ToolCalls) == 0 {
				fmt.Println(ui.GetStatusBar("● high · /effort"))
				break
			}

			// Handle tool calls
			for _, tc := range resp.ToolCalls {
				fmt.Printf(ui.UiMuted("🛠️  running %s...\n"), tc.Name)
				var content string
				if tc.Name == "bash" {
					fmt.Printf(ui.UiMuted("Command: %s\n"), string(tc.Arguments))
					fmt.Print("Approve? (y/N): ")
					approve := prompt.Input(fmt.Sprintf("%s (y/N): ", string(tc.Arguments)), func(d prompt.Document) []prompt.Suggest { return []prompt.Suggest{} })
					if strings.ToLower(strings.TrimSpace(approve)) == "y" {
						content = tools.ExecuteBash(ctx, tc.Arguments)
					} else {
						content = "User denied execution."
					}
				} else if tc.Name == "read_file" {
					content = tools.ExecuteReadFile(ctx, tc.Arguments)
				} else if tc.Name == "write_file" {
					content = tools.ExecuteWriteFile(ctx, tc.Arguments)
				} else {
					content = fmt.Sprintf("Error: Unknown tool %s", tc.Name)
				}
				history = append(history, provider.Message{
					Role:       provider.RoleTool,
					Content:    content,
					Name:       tc.Name,
					ToolCallID: tc.ID,
				})
			}
		}
	}
	return nil
}
