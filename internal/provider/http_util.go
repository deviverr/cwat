package provider

import (
	"fmt"
	"io"
	"strings"
)

const maxProviderResponseBytes int64 = 4 * 1024 * 1024

func readLimitedBody(r io.Reader, maxBytes int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > maxBytes {
		return nil, fmt.Errorf("response exceeds %d bytes", maxBytes)
	}
	return b, nil
}

func truncateForError(b []byte, maxChars int) string {
	s := strings.TrimSpace(string(b))
	if s == "" {
		return ""
	}
	if maxChars <= 0 || len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "..."
}
