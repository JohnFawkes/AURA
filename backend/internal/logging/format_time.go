package logging

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

// getLogTime retrieves the current time formatted as "2006/01/02 15:04:05"
// in the America/New_York timezone. It returns two strings: the formatted
// time with a grey color applied and the plain formatted time without color.
//
// Returns:
//   - colored (string): The formatted time string with grey color applied.
//   - formattedDT (string): The plain formatted time string without color.
func getLogTime() (string, string) {
	// Get the current time in the format "2006/01/02 15:04:05" for America/New_York
	dt := time.Now().Format("2006/01/02 15:04:05")

	// Change time zone to America/New_York
	loc, _ := time.LoadLocation("America/New_York")
	tz, _ := time.ParseInLocation("2006/01/02 15:04:05", dt, loc)

	// Format the time in the format "2006/01/02 15:04:05" for America/New_York
	formattedDT := tz.Format("2006/01/02 15:04:05")

	formattedDT = fixStringLength(formattedDT, TimeLength)
	colored := grey(formattedDT)

	// Return the colored time and the non-colored time
	return colored, formattedDT
}

// getLogTimeElapsed formats and returns the elapsed time in both colored and non-colored formats.
//
// Parameters:
//   - elapsed: A time.Duration value representing the elapsed time to be formatted.
//
// Returns:
//   - string: The elapsed time formatted with color for terminal output.
//   - string: The elapsed time formatted without color for plain text usage.
func getLogTimeElapsed(elapsed time.Duration) (string, string) {

	// Format the elapsed time
	formatted := formatElapsed(elapsed)
	formatted = fixStringLength(formatted, ByteAndElapsedLength)

	colored := color.CyanString(formatted)

	// Return the colored elapsed time and the non-colored elapsed time
	return colored, formatted
}

// formatElapsed formats a given time.Duration into a string with a fixed width of 5 digits
// for the numeric part, ensuring proper alignment and readability. The formatted string
// includes the integer part, two decimal places, and the unit (e.g., "ms", "s").
//
// The function ensures:
// - The integer part is right-aligned to 3 characters, padded with spaces if necessary.
// - The decimal part always has 2 digits, padded with zeros if missing.
// - The unit is preserved and appended at the end.
//
// Example:
//
//	Input: 12345 * time.Microsecond
//	Output: " 12.34 ms"
//
//	Input: 5 * time.Millisecond
//	Output: "  5.00 ms"
//
// Parameters:
// - elapsed: The time.Duration to format.
//
// Returns:
// - A formatted string representation of the elapsed time.
func formatElapsed(elapsed time.Duration) string {
	// Format elapsed time to ensure it is 5 digits total
	elapsedStr := elapsed.String()

	// Use regex to extract digits, decimal places, and characters
	re := regexp.MustCompile(`(\d+)(\.\d+)?([^\d]+)`)
	matches := re.FindStringSubmatch(elapsedStr)

	intPart := matches[1]
	decimalPart := matches[2]
	chars := matches[3]

	// Remove the decimal point from the decimal part
	decimalPart = strings.Replace(decimalPart, ".", "", 1)

	parsedIntPart, err := strconv.Atoi(intPart)
	if err != nil {
		intPart = "000"
	}

	if parsedIntPart < 100 {
		if len(intPart) == 1 {
			intPart = "  " + intPart
		} else if len(intPart) == 2 {
			intPart = " " + intPart
		}
	}

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
