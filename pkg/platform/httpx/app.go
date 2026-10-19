package httpx

import "strings"

// AppFromPath extracts the application segment from "/api/v1/{app}/...".
// Shared segments such as "auth" and "health" return an empty string.
func AppFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 || parts[0] != "api" || parts[1] != "v1" {
		return ""
	}
	seg := parts[2]
	if seg == "auth" || seg == "health" {
		return ""
	}
	return seg
}
