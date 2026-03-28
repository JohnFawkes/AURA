package mediux

import (
	"aura/cache"
	"aura/logging"
	"aura/models"
	"aura/utils/httpx"
	"context"
	_ "embed"
)

//go:embed gen_items_with_sets.graphql
var queryItemsWithSets string

func PreLoadMediuxItemsWithSets(ctx context.Context) {
	ctx, logAction := logging.AddSubActionToContext(ctx, "Preloading MediUX Items with Sets", logging.LevelTrace)
	defer logAction.Complete()

	// Send the request
	URL := "https://api.mediux.io/v1/list/content_ids"
	resp, respBody, Err := makeRequest(ctx, URL, "GET", nil, "", false)
	if Err.Message != "" {
		logging.LOGGER.Error().Timestamp().Msgf("Failed to make request to MediUX API for items with sets: %s", Err.Message)
		return
	}
	defer resp.Body.Close()

	currentCount := cache.MediuxItems.GetCountMediuxItems()
	logAction.AppendResult("current_count", currentCount)

	var mediuxResp models.MediuxContentIdsResponse

	// Decode the response
	Err = httpx.DecodeResponseToJSON(ctx, respBody, &mediuxResp, "MediUX Items with Sets Response Decoding")
	if Err.Message != "" {
		logging.LOGGER.Error().Timestamp().Msgf("Failed to decode MediUX items with sets response: %s", Err.Message)
		return
	}

	itemCount := len(mediuxResp.Items)
	logAction.AppendResult("new_count", itemCount)

	cache.MediuxItems.StoreMediuxItems(mediuxResp.Items)
	logEvent := logging.LOGGER.Info().Timestamp()
	if currentCount > 0 {
		logEvent = logEvent.Int("current_count", currentCount)
	}
	logEvent.Int("new_count", itemCount).Msg("Loaded MediUX items with sets into cache")
}
