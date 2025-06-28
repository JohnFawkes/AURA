package utils

import (
	"aura/internal/logging"
	"fmt"
)

func PrintBanner(Version string, Author string, License string, port int) {

	logging.LOG.NoTime("┌────────────────────────────────┐\n")
	logging.LOG.NoTime(formatBanner("aura"))
	logging.LOG.NoTime(formatBanner("App Version: " + Version))
	logging.LOG.NoTime(formatBanner("Author: " + Author))
	logging.LOG.NoTime(formatBanner("License: " + License))
	logging.LOG.NoTime(formatBanner(fmt.Sprintf("API Port: %d", port)))
	logging.LOG.NoTime("└────────────────────────────────┘\n")

}

func formatBanner(text string) string {
	// Format the text to fit nicely in the banner
	// Pad the text to fit within the banner width
	if len(text) > 30 {
		text = text[:30]
	}
	for len(text) < 30 {
		text += " "
	}

	// If the text starts with ┌ or └, don't add the " │ " prefix
	if len(text) > 0 && (rune(text[0]) == '┌' || rune(text[0]) == '└') {
		return fmt.Sprintf(" %s\n", text)
	}
	// Return the formatted text
	return fmt.Sprintf("│ %s │\n", text)
}
