package httpx

import "strings"

func MakeAuthHeader(key string, value string) map[string]string {
	headers := make(map[string]string)
	if strings.ToLower(key) == "authorization" && !strings.HasPrefix(strings.ToLower(value), "bearer ") {
		value = "Bearer " + value
	}
	headers[key] = value
	return headers
}
