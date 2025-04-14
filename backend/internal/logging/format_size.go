package logging

import (
	"fmt"

	"github.com/fatih/color"
)

// getLogBytes formats and colors a given byte size for logging purposes.
//
// Parameters:
//   - bytes: The size in bytes to be formatted and colored.
//
// Returns:
//   - A string representing the colored and formatted byte size.
//   - A string representing the non-colored, formatted byte size.
func getLogBytes(bytes int) (string, string) {

	// Format the bytes written
	formatted := formatBytes(bytes)
	formatted = fixStringLength(formatted, ByteAndElapsedLength)

	// Color the bytes written
	colored := colorBytes(formatted)

	// Return the colored bytes and the non-colored bytes
	return colored, formatted
}

// formatBytes converts a size in bytes to a human-readable string representation
// with appropriate units (B, KB, MB, GB). The function ensures that the output
// is formatted to two decimal places for values less than 100 and adjusts the
// unit dynamically based on the size of the input.
//
// Parameters:
//   - bytes: The size in bytes to be formatted.
//
// Returns:
//
//	A string representing the formatted size with the appropriate unit.
func formatBytes(bytes int) string {
	unit := ""
	value := float64(bytes)

	switch {
	case value >= 1<<30:
		unit = " GB"
		value /= 1 << 30
	case value >= 1<<20:
		unit = " MB"
		value /= 1 << 20
	case value >= 1<<10:
		unit = " KB"
		value /= 1 << 10
	default:
		unit = " B"
	}

	if value == 0 {
		str := "  0.00 B"
		return str
	}

	if value > 0 && value < 100 {
		return fmt.Sprintf(" %.2f%s", value, unit)
	}

	return fmt.Sprintf("%.2f%s", value, unit)
}

// colorBytes takes a string representing a size (e.g., "10GB", "512MB", "128KB")
// and returns the string formatted with a color based on the size unit.
// - "GB" is formatted in red.
// - "MB" is formatted in yellow.
// - "KB" is formatted in green.
// - Any other unit is formatted in blue.
func colorBytes(bytes string) string {
	end := bytes[len(bytes)-2:]
	switch {
	case end == "GB":
		return color.RedString(bytes)
	case end == "MB":
		return color.YellowString(bytes)
	case end == "KB":
		return color.GreenString(bytes)
	default:
		return color.BlueString(bytes)
	}
}
