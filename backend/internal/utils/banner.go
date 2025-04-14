package utils

import "fmt"

func PrintBanner(Version string, Author string, License string, port int, logLevel string) {
	fmt.Printf(("┌────────────────────────────────┐\n"))
	fmt.Print(formatBanner("Poster Setter"))
	fmt.Print(formatBanner("Version: " + Version))
	fmt.Print(formatBanner("Author: " + Author))
	fmt.Print(formatBanner("License: " + License))
	fmt.Print(formatBanner(fmt.Sprintf("Port: %d", port)))
	fmt.Print(formatBanner(fmt.Sprintf("Log Level: %s", logLevel)))
	fmt.Print(("└────────────────────────────────┘\n"))
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
