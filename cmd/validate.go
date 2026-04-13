package cmd

import (
	"fmt"
	"strings"
	"unicode"
)

// validateResourceID checks an ID for agent hallucination patterns.
func validateResourceID(id, label string) error {
	if id == "" {
		return fmt.Errorf("%s cannot be empty", label)
	}
	if strings.Contains(id, "..") {
		return fmt.Errorf("invalid %s: path traversal detected", label)
	}
	lower := strings.ToLower(id)
	if strings.Contains(lower, "%2e") || strings.Contains(lower, "%2f") {
		return fmt.Errorf("invalid %s: encoded path characters detected", label)
	}
	if strings.ContainsAny(id, "?#&=") {
		return fmt.Errorf("invalid %s: query parameters not allowed in IDs", label)
	}
	if strings.ContainsAny(id, "/\\") {
		return fmt.Errorf("invalid %s: path separators not allowed in IDs", label)
	}
	for _, r := range id {
		if unicode.IsControl(r) {
			return fmt.Errorf("invalid %s: control characters not allowed", label)
		}
	}
	return nil
}
