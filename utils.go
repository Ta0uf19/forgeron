package forgeron

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// pascalizeHeaders converts HTTP/2 headers to their proper case format
func pascalizeHeaders(headers map[string]string) map[string]string {
	h := make(map[string]string)
	for k, v := range headers {
		switch strings.ToLower(k) {
		// note: sec-ch- / client hints headers are in lowercase deliberately.
		case "user-agent", "accept-language", "accept-encoding", "accept",
			"content-type", "content-length", "connection", "host", "referer",
			"origin", "cache-control", "pragma", "upgrade-insecure-requests",
			"sec-fetch-mode", "sec-fetch-dest", "sec-fetch-site", "sec-fetch-user":
			h[pascalizeKey(k)] = v
		default:
			h[k] = v
		}
	}
	return h
}

// pascalizeKey converts a string to PascalCase
func pascalizeKey(key string) string {
	return cases.Title(language.Und).String(key)
}

// atoi converts a string to an integer
func atoi(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// parseMap parses a JSON string into a map
func parseMap(s string) map[string]string {
	var m map[string]string
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return make(map[string]string)
	}
	return m
}

// parseString parses a string and returns a pointer to it
func parseStringPtr(s string) *string {
	if s == "" || s == "null" {
		return nil
	}
	return &s
}

// parseInt parses a string and returns an integer
func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// parseIntPtr parses a string and returns a pointer to an integer
func parseIntPtr(s string) *int {
	if s == "" || s == "null" {
		return nil
	}
	i, _ := strconv.Atoi(s)
	return &i
}
