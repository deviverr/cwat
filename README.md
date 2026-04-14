# ?? CWAT (Code With AI Terminal)

**Ai console client written in Go.**

CWAT is a clean-room, multi-provider, agentic CLI tool tailored for developers. Built entirely in Go, it delivers a lightning-fast, highly polished terminal experience with built-in tools that allow the AI to safely interact with your local environment.

## ? Features

- **?? Lightning Fast:** Compiled to a single Go binary. No massive `node_modules` or runtime latency. Boot up is instantaneous.
- **?? Agentic Tools:** The assistant isn't blind. It can natively use tools to:
  - Read files and navigate your codebase.
  - Write and replace file content.
  - Execute `bash`/`powershell` commands (with a secure **[y/N]** human-in-the-loop approval prompt).
- **?? Beautiful TUI:** Powered by `charmbracelet/lipgloss` and `glamour`. Output isn't just text; it's fully rendered Markdown with syntax-highlighted code blocks, responsive flex-box dashboards, and dynamic themes.
- **?? Interactive Repl:** A robust, arrow-key navigable autocomplete menu for built-in slash commands (powered by `go-prompt`).
- **?? Multi-Provider Support:** Easily hook up to OpenAI, Anthropic, or Qwen models natively.

## ?? Installation

Make sure you have [Go](https://golang.org/dl/) installed.

```bash
# Clone the repository
git clone https://github.com/deviverr/cwat.git
cd cwat

# Install the binary
go install ./cmd/cwat
```

## ?? Getting Started

To initialize CWAT for the first time, run the setup wizard to configure your preferred AI provider and API keys:

```bash
cwat setup
```

Launch the interactive chat loop:

```bash
cwat chat
```

*Tip: You can pass specific profiles or override the model directly:*
```bash
cwat chat --model qwen3-14b
```

## ?? Slash Commands

Type `/` in the interactive prompt to open the dropdown menu:
- `/help` - Show all commands
- `/clear` - Clear conversation history
- `/theme` - Cycle through TUI color themes
- `/cost` - Show estimated session cost
- `/exit` - Quit CWAT gracefully

## ??? Security (Human-in-the-Loop)

CWAT gives AI models immense power to edit code and run shell commands. By default, **any destructive action or shell command execution requires explicit user approval**. When the AI attempts to run a bash script, the prompt will suspend and ask you: `Approve? (y/N)`.

## ??? Built With
- [go-prompt](https://github.com/c-bata/go-prompt) - Advanced interactive CLI prompts.
- [glamour](https://github.com/charmbracelet/glamour) - Native terminal markdown rendering.
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Beautiful layout styling.

---
*Built for the command line. Built for developers.*

