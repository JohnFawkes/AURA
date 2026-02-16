package mediux

import (
	"aura/logging"
	"fmt"
	"net/http"
	"time"
)

var MediuxSiteLink string = ""

func init() {
	MediuxSiteLink = ""
}

func CheckSiteLinkAvailability() {
	itemType := "movie"
	tmdbID := "550"

	mainURL := "https://mediux.io"
	backupURL := "https://mediux.pro"

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Head(fmt.Sprintf("%s/%s/%s", mainURL, itemType, tmdbID))
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil || resp.StatusCode != http.StatusOK {
		resp, err = client.Head(fmt.Sprintf("%s/%ss/%s", backupURL, itemType, tmdbID))
		if resp != nil {
			defer resp.Body.Close()
		}
		if err != nil || resp.StatusCode != http.StatusOK {
			return
		} else {
			MediuxSiteLink = backupURL
			logging.LOGGER.Info().Timestamp().Msg("Mediux Site Link set to backup URL: " + MediuxSiteLink)
			return
		}
	}
	MediuxSiteLink = mainURL
	logging.LOGGER.Info().Timestamp().Msg("Mediux Site Link set to main URL: " + MediuxSiteLink)
}
