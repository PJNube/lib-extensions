package naming

import (
	"regexp"
	"strings"
)

var cleanRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// Sanitize removes invalid characters and lowercases the result.
// Keeps only [a-zA-Z0-9_], then enforces length 2–32.
func Sanitize(s string) string {
	s = cleanRe.ReplaceAllString(s, "")
	s = strings.ToLower(s)
	if len(s) < 2 {
		return "na"
	}
	if len(s) > 32 {
		s = s[:32]
	}
	return s
}

func SanitizeProfile(s string) string {
	s = strings.ToLower(s)
	if len(s) < 2 {
		return "na"
	}
	if len(s) > 2 {
		s = s[:2]
	}
	return s
}
