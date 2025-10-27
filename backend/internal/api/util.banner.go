package api

import (
	"aura/internal/logging"
	"fmt"
)

func Util_Banner_Print(Version string, Author string, License string, port int) {

	logging.LOG.NoTime("┌────────────────────────────────┐\n")
	logging.LOG.NoTime(Util_Banner_Format("aura"))
	logging.LOG.NoTime(Util_Banner_Format("App Version: " + Version))
	logging.LOG.NoTime(Util_Banner_Format("Author: " + Author))
	logging.LOG.NoTime(Util_Banner_Format("License: " + License))
	logging.LOG.NoTime(Util_Banner_Format(fmt.Sprintf("API Port: %d", port)))
	logging.LOG.NoTime("└────────────────────────────────┘\n")

}

func Util_Banner_Format(text string) string {
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
