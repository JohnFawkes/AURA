package logging

import (
	"strings"

	"github.com/fatih/color"
)

// fixStringLength ensures that a given string has a specific length by either
// appending spaces to the end of the string if it is shorter than the desired
// length, or returning the string as-is if it meets or exceeds the length.
//
// Parameters:
//   - str: The input string to be adjusted.
//   - length: The desired length of the string.
//
// Returns:
//
//	A string of the specified length, padded with spaces if necessary.
func fixStringLength(str string, length int) string {
	// If the string is shorter than the length, add spaces to the end
	if len(str) < length {
		str += strings.Repeat(" ", length-len(str))
	}
	return str
}

// colorLevel applies color formatting to a log level string based on its severity.
//
// Parameters:
//   - level: A string representing the log level (e.g., "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL").
//
// Behavior:
//   - If the level contains "TRACE", it is colored magenta.
//   - If the level is "DEBUG", it is colored cyan.
//   - If the level is "INFO", it is colored green.
//   - If the level is "WARN", it is colored yellow.
//   - If the level is "ERROR", it is colored red.
//   - If the level is "FATAL", it is colored bright red.
//   - For any other level, the original string is returned without coloring.
//
// Returns:
//   - A string with the appropriate color formatting applied.
func colorLevel(level string) string {
	// Remove spaces from the level string for comparison
	lvl := strings.ReplaceAll(level, " ", "")

	switch lvl {
	case "TRACE":
		return color.MagentaString(level)
	case "DEBUG":
		return color.CyanString(level)
	case "INFO":
		return color.GreenString(level)
	case "WARN":
		return color.YellowString(level)
	case "ERROR":
		return color.RedString(level)
	case "FATAL":
		return color.HiRedString(level)
	default:
		return level
	}
}
