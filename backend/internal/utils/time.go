package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Calculate the elapsed time since the start of the task
func ElapsedTime(start time.Time) string {
	elapsed := time.Since(start)

	// Convert the elapsed time to a string
	elapsedStr := elapsed.String()

	// Use regex to extract digits, decimal places, and characters
	re := regexp.MustCompile(`(\d+)(\.\d+)?([^\d]+)`)
	matches := re.FindStringSubmatch(elapsedStr)

	intPart := matches[1]
	decimalPart := matches[2]
	chars := matches[3]

	// Remove the decimal point from the decimal part
	decimalPart = strings.Replace(decimalPart, ".", "", 1)

	if len(decimalPart) == 0 || decimalPart == "" {
		decimalPart = "00"

	}

	// Combine integer part, decimal places, and characters with right alignment
	if len(decimalPart) < 2 {
		decimalPart += strings.Repeat("0", 2)
	}

	formattedElapsed := fmt.Sprintf("%s.%s %s", intPart, decimalPart[:2], chars)
	return formattedElapsed
}
