package runtime

import (
	"fmt"
	"strings"

	"github.com/cwat-code/cwat-go-client/internal/config"
	"github.com/cwat-code/cwat-go-client/internal/provider"
)

// BuildClient creates the concrete provider client for a profile.
func BuildClient(profile config.Profile, apiKey string) (provider.Client, error) {
	kind := strings.ToLower(strings.TrimSpace(profile.Provider))
	switch kind {
	case "anthropic":
		return provider.NewAnthropicClient(profile.BaseURL, apiKey, profile.AnthropicVersion), nil
	case "openai":
		return provider.NewOpenAIClient(profile.BaseURL, apiKey, "openai"), nil
	case "openai_compatible", "openai-compatible", "openai_compat":
		return provider.NewOpenAIClient(profile.BaseURL, apiKey, "openai_compatible"), nil
	default:
		return nil, fmt.Errorf("unsupported provider %q", profile.Provider)
	}
}
