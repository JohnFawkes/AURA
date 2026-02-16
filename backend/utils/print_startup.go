package utils

import (
	"aura/logging"
)

// Print application startup details
func PrintAppStartUpDetails(APP_VERSION, AUTHOR, LICENSE string, APP_PORT int, APP_NAME string) {
	logging.LOGGER.Info().
		Timestamp().
		Str("version", APP_VERSION).
		Str("author", AUTHOR).
		Str("license", LICENSE).
		Int("port", APP_PORT).
		Msgf("Started %s", APP_NAME)
}
