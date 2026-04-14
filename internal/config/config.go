package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ConfigPathEnv overrides the config file location.
	ConfigPathEnv = "CWAT_CONFIG"
)

// Config is the top-level cwat config structure.
type Config struct {
	ActiveProfile string             `json:"active_profile"`
	Profiles      map[string]Profile `json:"profiles"`
	Setup         SetupPreferences   `json:"setup,omitempty"`
}

// SetupPreferences stores setup wizard choices and onboarding approvals.
type SetupPreferences struct {
	Completed                  bool                  `json:"completed,omitempty"`
	Theme                      string                `json:"theme,omitempty"`
	SyntaxHighlightingDisabled bool                  `json:"syntax_highlighting_disabled,omitempty"`
	Permissions                PermissionPreferences `json:"permissions,omitempty"`
	ApprovedAPIKeyHints        []string              `json:"approved_api_key_hints,omitempty"`
}

// PermissionPreferences stores coarse trust choices for local tools.
type PermissionPreferences struct {
	AllowBash  bool `json:"allow_bash,omitempty"`
	AllowMCP   bool `json:"allow_mcp,omitempty"`
	AllowHooks bool `json:"allow_hooks,omitempty"`
	AllowEnv   bool `json:"allow_env,omitempty"`
}

// Profile defines one provider/model endpoint.
type Profile struct {
	Provider         string  `json:"provider"`
	Model            string  `json:"model"`
	APIKeyEnv        string  `json:"api_key_env,omitempty"`
	APIKey           string  `json:"api_key,omitempty"`
	BaseURL          string  `json:"base_url,omitempty"`
	AnthropicVersion string  `json:"anthropic_version,omitempty"`
	MaxTokens        int     `json:"max_tokens,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"`
}

// DefaultPath returns the default config path (~/.cwat/config.json).
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".cwat", "config.json"), nil
}

// ResolvePath resolves config path from argument, env, or default.
func ResolvePath(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return explicit, nil
	}
	if v := strings.TrimSpace(os.Getenv(ConfigPathEnv)); v != "" {
		return v, nil
	}
	return DefaultPath()
}

// Load loads and validates config.
func Load(explicit string) (Config, string, error) {
	path, err := ResolvePath(explicit)
	if err != nil {
		return Config{}, "", err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, path, fmt.Errorf("config not found at %s (run 'cwat init')", path)
		}
		return Config{}, path, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, path, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, path, err
	}
	return cfg, path, nil
}

// Validate checks required fields.
func (c Config) Validate() error {
	if len(c.Profiles) == 0 {
		return errors.New("config has no profiles")
	}
	active := strings.TrimSpace(c.ActiveProfile)
	if active == "" {
		return errors.New("active_profile is required")
	}
	if _, ok := c.Profiles[active]; !ok {
		return fmt.Errorf("active_profile %q does not exist", active)
	}
	if c.Setup.Theme != "" {
		theme := strings.ToLower(strings.TrimSpace(c.Setup.Theme))
		if theme != "dark" && theme != "light" {
			return fmt.Errorf("setup.theme must be 'dark' or 'light'")
		}
	}
	for name, p := range c.Profiles {
		if strings.TrimSpace(name) == "" {
			return errors.New("profile names must not be empty")
		}
		if strings.TrimSpace(p.Provider) == "" {
			return fmt.Errorf("profiles.%s.provider is required", name)
		}
		if strings.TrimSpace(p.Model) == "" {
			return fmt.Errorf("profiles.%s.model is required", name)
		}
		if strings.TrimSpace(p.APIKeyEnv) == "" && strings.TrimSpace(p.APIKey) == "" {
			return fmt.Errorf("profiles.%s.api_key_env or profiles.%[1]s.api_key is required", name)
		}

		providerKind := strings.ToLower(strings.TrimSpace(p.Provider))
		switch providerKind {
		case "anthropic", "openai", "openai_compatible", "openai-compatible", "openai_compat":
		default:
			return fmt.Errorf("profiles.%s.provider %q is not supported", name, p.Provider)
		}

		if p.MaxTokens < 0 {
			return fmt.Errorf("profiles.%s.max_tokens must be >= 0", name)
		}
		if p.Temperature < 0 || p.Temperature > 2 {
			return fmt.Errorf("profiles.%s.temperature must be in range [0, 2]", name)
		}

		if (providerKind == "openai_compatible" || providerKind == "openai-compatible" || providerKind == "openai_compat") && strings.TrimSpace(p.BaseURL) == "" {
			return fmt.Errorf("profiles.%s.base_url is required for openai_compatible providers", name)
		}
	}
	return nil
}

// Save writes config to the resolved config path.
func Save(explicit string, cfg Config) (string, error) {
	if err := cfg.Validate(); err != nil {
		return "", err
	}

	path, err := ResolvePath(explicit)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("render config: %w", err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}
	return path, nil
}

// ResolveProfile picks explicit profile or active profile.
func (c Config) ResolveProfile(explicit string) (string, Profile, error) {
	name := strings.TrimSpace(explicit)
	if name == "" {
		name = c.ActiveProfile
	}
	p, ok := c.Profiles[name]
	if !ok {
		return "", Profile{}, fmt.Errorf("profile %q not found", name)
	}
	return name, p, nil
}

// ReadAPIKey reads API key from profile APIKeyEnv.
func (p Profile) ReadAPIKey() (string, error) {
	if strings.TrimSpace(p.APIKey) != "" {
		return strings.TrimSpace(p.APIKey), nil
	}
	if p.APIKeyEnv == "" {
		return "", fmt.Errorf("no APIKey or APIKeyEnv configured")
	}
	k := strings.TrimSpace(os.Getenv(p.APIKeyEnv))
	if k == "" {
		return "", fmt.Errorf("missing API key env %s", p.APIKeyEnv)
	}
	return k, nil
}

// SaveExample writes an example config to the resolved path.
func SaveExample(explicit string) (string, error) {
	path, err := ResolvePath(explicit)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		return path, fmt.Errorf("config already exists at %s", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return path, fmt.Errorf("check config path: %w", err)
	}

	b, err := json.MarshalIndent(ExampleConfig(), "", "  ")
	if err != nil {
		return "", fmt.Errorf("render config: %w", err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}
	return path, nil
}

// ExampleConfig returns a starter config with three profiles.
func ExampleConfig() Config {
	return Config{
		ActiveProfile: "anthropic",
		Profiles: map[string]Profile{
			"anthropic": {
				Provider:         "anthropic",
				Model:            "CWAT-sonnet-4-5",
				APIKeyEnv:        "ANTHROPIC_API_KEY",
				AnthropicVersion: "2023-06-01",
				MaxTokens:        1200,
				Temperature:      0.2,
			},
			"openai": {
				Provider:    "openai",
				Model:       "gpt-4.1-mini",
				APIKeyEnv:   "OPENAI_API_KEY",
				BaseURL:     "https://api.openai.com/v1",
				MaxTokens:   1200,
				Temperature: 0.2,
			},
			"own": {
				Provider:    "openai_compatible",
				Model:       "my-model",
				APIKeyEnv:   "MY_PROVIDER_API_KEY",
				BaseURL:     "https://my-provider.example/v1",
				MaxTokens:   1200,
				Temperature: 0.2,
			},
			"qwen3": {
				Provider:    "openai",
				Model:       "qwen3-14b",
				APIKey:      "sk-47BhcY3Seigp5WmpwzDp8w",
				BaseURL:     "https://gw-lite.777wt.net/v1",
				MaxTokens:   4000,
				Temperature: 0.7,
			},
		},
	}
}
