package masking

import (
	"regexp"
)

func Masking_IpAddress(logContent string) string {
	patterns := map[string]string{
		`\b\d{1,3}(\.\d{1,3}){3}\b`: "***REDACTED_IP***", // IP addresses
	}

	for pattern, replacement := range patterns {
		re := regexp.MustCompile(pattern)
		logContent = re.ReplaceAllString(logContent, replacement)
	}
	return logContent
}
