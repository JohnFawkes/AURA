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

	// Send the GraphQL request
	resp, Err := makeGraphQLRequest(ctx, MediuxGraphQLQueryBody{
		Query:     queryItemsWithSets,
		Variables: map[string]any{},
		QueryName: "getItemsWithSets",
	})
	if Err.Message != "" {
		return
	}

	currentCount := cache.MediuxItems.GetCountMediuxItems()
	logAction.AppendResult("current_count", currentCount)

	var mediuxResp models.MediuxContentIdsResponse

	// Decode the response
	Err = httpx.DecodeResponseToJSON(ctx, resp.Body(), &mediuxResp, "MediUX Items with Sets Response Decoding")
	if Err.Message != "" {
		return
	}

	itemCount := len(mediuxResp.Data.ContentIds)
	logAction.AppendResult("new_count", itemCount)

	cache.MediuxItems.StoreMediuxItems(mediuxResp.Data.ContentIds)
	logEvent := logging.LOGGER.Info().Timestamp()
	if currentCount > 0 {
		logEvent = logEvent.Int("current_count", currentCount)
	}
	logEvent.Int("new_count", itemCount).Msg("Loaded MediUX items with sets into cache")
}
