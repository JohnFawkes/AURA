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

	var mediuxResp models.MediuxContentIdsResponse

	// Decode the response
	Err = httpx.DecodeResponseToJSON(ctx, resp.Body(), &mediuxResp, "MediUX Items with Sets Response Decoding")
	if Err.Message != "" {
		return
	}

	itemCount := len(mediuxResp.Data.ContentIds)
	logAction.AppendResult("items_with_sets_count", itemCount)

	cache.MediuxItems.StoreMediuxItems(mediuxResp.Data.ContentIds)
}
